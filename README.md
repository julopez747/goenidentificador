# goeniidentificador

Generación del metadato identificador para documento y expediente electrónico

Levanta un servicio http por el puerto 8080 con las siguientes URLs

  /ping 
                devuelve pong
  /healthz
                devuelve OK si el servicio es capaz de responder peticiones
  /eniidentificador/documento/{unidad}/{anio}/{tipo}
                recibe:
                  - unidad: código dir3 de la unidad productora (9 dígitos)
                  - anio: el año para el documento (4 dígitos)
                  - tipo: código del tipo de documento (2 dígitos)
                devuelve:
                  - un status 200 con un JSON con el campo identificador del siguiente identificador de la unidad y año
                  - un status 400 si alguno de los parámetros no tiene los valores correctos
                  - un status 500 en caso de problemas con la conexión a la base de datos

  /eniidentificador/documento/{unidad}/{anio}/{serie}
                recibe:
                  - unidad: código dir3 de la unidad productora (9 dígitos)
                  - anio: el año para el documento (4 dígitos)
                  - serie: información de la clasificación lǵica, serie y subserie documental (8 dígitos)
                devuelve:
                  - un status 200 con un JSON con el campo identificador del siguiente identificador de la unidad y año
                  - un status 400 si alguno de los parámetros no tiene los valores correctos
                  - un status 500 en caso de problemas con la conexión a la base de datos

## Base de datos

Se hace uso del SGBD postgresql, al arrancar crea, si no existen, las tablas y resto de elementos necesarios para la ejecución del proyecto

En concreto se crea:

- Tabla: goeniidentificador que almacena para cada expediente o documento, unidad y año el siguiente número de identificador a generar.

## Variables de entorno

El proyecto usa dos variables de entorno:

 DATABASE_CONN_STRING con la cadena de conexión a la base de datos postgresql
 DATABASE_SCHEMA esquema de la base de datos en que se encuentran las tablas

