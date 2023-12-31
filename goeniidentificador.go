package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	_ "github.com/lib/pq"
)

/*
Variables globales
*/
var Database *sql.DB        // conexión a la base de datos
var Esquema string          // valor recuperado de las variables de entorno con el nombre del esquema
var CadenaDeConexion string // valor recuperado de las variables de entorno con la cadena de conexión

const NombreTabla = "goeniidentificador" // nombre de la tabla de este proyecto

/*
Función principal que recupera los datos del SO, crea las tablas si hacen falta e inicia el servicio
*/
func main() {
	var err error

	// Recupera las variables del entorno
	CadenaDeConexion = os.Getenv("DATABASE_CONN_STRING")
	Esquema = os.Getenv("DATABASE_SCHEMA")

	// Obtiene una conexión a la base de datos
	Database, err = sql.Open("postgres", CadenaDeConexion)
	if err != nil {
		log.Panic(err)
	}

	// Si no existe el esquema lo crea
	err = generaEsquema()
	if err != nil {
		log.Panic(err)
	}

	// Ejecuta el servicio
	log.Println("iniciando el servicio")
	inicializaRouter()

	Database.Close()

}

/*
Inicializa el router y espera las conexiones
*/
func inicializaRouter() {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(render.SetContentType(render.ContentTypeJSON))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Ops. Nothing here"))
	})

	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		// go mantiene el pool de conexiones a la base de datos, por lo tanto OK
		w.Write([]byte("OK"))
	})

	router.Route("/eniidentificador", func(r chi.Router) {

		r.Route("/documento/{unidad}/{anio}/{serie}", func(r chi.Router) {
			r.Get("/", RecuperaIdentificadorDocumento)
		})

		r.Route("/expediente/{unidad}/{anio}/{serie}", func(r chi.Router) {
			r.Get("/", RecuperaIdentificadorExpediente)
		})

	})

	http.ListenAndServe(":8080", router)
}

/*
Función que recupera el identificador del documento basado en el path
*/
func RecuperaIdentificadorDocumento(w http.ResponseWriter, r *http.Request) {
	RecuperaIdentificador(w, r, "D")
}

/*
Función que recupera el identificador del expediente basado en el path
*/
func RecuperaIdentificadorExpediente(w http.ResponseWriter, r *http.Request) {
	RecuperaIdentificador(w, r, "E")
}

/*
Función que recupera el identificador basado en el path
*/
func RecuperaIdentificador(w http.ResponseWriter, r *http.Request, modo string) {
	var correcto bool = true // indica si los parámetros son correctos
	var anio int             // valor del año
	var identificador int    // valor del identificador asignado
	var serie string         // valor de la serie para genera el identificador de expediente y documento
	var err error            // error

	// recupera los parámetros
	unidad := chi.URLParam(r, "unidad")
	cadenaAnio := chi.URLParam(r, "anio")
	serie = chi.URLParam(r, "serie")

	// valida que se han definido los parámetros, y las longitudes (para año, unidad y serie)
	if correcto && (len(unidad) != 9 || len(cadenaAnio) != 4) {
		correcto = false
	}

	if correcto && ((modo == "E") && len(serie) != 8) {
		correcto = false
	}

	if correcto && ((modo == "D") && len(serie) != 2) {
		correcto = false
	}

	if correcto {
		anio, err = strconv.Atoi(cadenaAnio)
		correcto = err == nil
	}

	// Si es correcto, genera el identificador.
	// Nota: cuando el identificador es 0 y no existe error, es que acaba de crear la serie y
	// se tiene que volver a pedir el identificador (forma de que sea transacción ACID)
	if correcto {
		identificador = 0
		err = nil
		for identificador == 0 && err == nil {
			identificador, err = obtenerIdentificador(modo, unidad, anio)
			if err != nil {
				correcto = false
				log.Printf("Error: %v", err.Error())
				errorInterno := errors.New("error interno del servidor")
				respuesta := ErrResponse{
					Err:            errorInterno,
					HTTPStatusCode: 500,
					StatusText:     "Internal server error.",
					ErrorText:      errorInterno.Error(),
				}
				render.Render(w, r, &respuesta)
			}
		}
		correcto = err == nil
	}

	// Envía el resultado 200 OK o un 400 si los datos son incorrectos
	if correcto {
		identificadorNormailizado := fmt.Sprintf("%v", identificador)
		if modo == "E" {
			padding := strings.Repeat("0", 21-len(identificadorNormailizado))
			identificadorNormailizado = fmt.Sprintf(
				"ES_%v_%v_EXP_%v_%v%v",
				unidad,
				anio,
				serie,
				padding,
				identificadorNormailizado,
			)
		} else if modo == "D" {
			padding := strings.Repeat("0", 28-len(identificadorNormailizado))
			identificadorNormailizado = fmt.Sprintf(
				"ES_%v_%v_%v%v%v",
				unidad,
				anio,
				serie,
				padding,
				identificadorNormailizado,
			)
		}
		respuesta := Respuesta{Identificador: identificadorNormailizado}
		render.Render(w, r, &respuesta)
		log.Println(identificador)
	} else {
		errorPeticion := errors.New("invalid request")
		respuesta := ErrResponse{
			Err:            errorPeticion,
			HTTPStatusCode: 400,
			StatusText:     "Invalid request.",
			ErrorText:      errorPeticion.Error(),
		}
		render.Render(w, r, &respuesta)
	}

}

/*
Estructrua de la respuesta, incluye el identificador
*/
type Respuesta struct {
	Identificador string `json:"identificador"`
}

/*
Función para generar la respuesta en JSON
*/
func (e *Respuesta) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, 200)
	return nil
}

/*
Estructura para la generación de errores
*/
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

/*
Función para generar el error en JSON
*/
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

/*
Crea las tablas necesarias para el proyecto si no existen.
*/
func generaEsquema() error {
	sqlExiste := "select count(1) from pg_tables where lower(schemaname) = lower($1) and lower(tablename) = $2"
	var cuenta int

	registroExiste := Database.QueryRow(sqlExiste, Esquema, NombreTabla)
	if err := registroExiste.Scan(&cuenta); err != nil {
		log.Println(err)
		return err
	}

	if cuenta == 0 {
		log.Println("no se han detectado las tablas del sistema, creando tablas")
		sqlCreaTabla := fmt.Sprintf(
			`create table %v.%v (
				dir3 varchar(9) not null,
				anio numeric(9) not null,
				tipo varchar(2) not null,
				indice numeric(12),
				primary key(dir3, anio, tipo)
			)`, Esquema, NombreTabla)

		_, err := Database.Exec(sqlCreaTabla)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
Obtiene el siguiente identificador para el tipo indicado de la unidad dir3 en el año indicado.
Si no existe la secuencia, la crea y devuelve como identificador 0 y no genera error
*/
func obtenerIdentificador(tipo string, dir3 string, anio int) (identificador int, err error) {
	identificador = 0
	sqlUpdate := fmt.Sprintf(
		`UPDATE %v.%v
		SET indice = indice + 1
		WHERE dir3 = $1 AND anio = $2 AND tipo = $3
		RETURNING indice`,
		Esquema,
		NombreTabla,
	)

	registro := Database.QueryRow(sqlUpdate, dir3, anio, tipo)
	if err := registro.Scan(&identificador); err != nil {
		if err == sql.ErrNoRows {
			sqlInsert := fmt.Sprintf(
				"INSERT INTO %v.%v (dir3, anio, tipo, indice) values($1, $2, $3, 0)",
				Esquema,
				NombreTabla,
			)
			_, _ = Database.Exec(sqlInsert, dir3, anio, tipo)
			return 0, nil
		}
		return 0, err
	}
	return
}
