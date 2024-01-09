# goeniidentificador

Generación del metadato identificador para documento y expediente electrónico

Levanta un servicio http por el puerto 8080 con las siguientes URLs

| Path | Método | Acción |
| ------ | ----- | ----- |
| /ping | GET | devuelve *pong* |
|/healthz| GET | devuelve *OK* si el servicio es capaz de responder peticiones|
|/eniidentificador/documento/{unidad}/{anio}/{tipo}| GET | Obtiene el identificador de un documento, [detalles](#obtener-identificador-de-documento)
|/eniidentificador/expediente/{unidad}/{anio}/{serie}| GET | Obtiene el identificador de un expediente, [detalles](#obtener-identificador-de-expediente)



## Obtener identificador de documento

Recibe:

| Campo | Descripcion | Tamaño |
| ----- | ------------| -------|
| **unidad** | código dir3 de la unidad productora | 9 dígitos |
|  **anio** | el año para el documento |4 dígitos |
|  **tipo** | código del tipo de documento | 2 dígitos |

Resultado:

| Estado HTTP | Descripción |
| ----------- | ----------- |
| 200 | JSON con el campo **identificador** del siguiente identificador de documento para la unidad y año |
| 400 | si alguno de los parámetros no tiene los valores correctos |
| 500 | si se produce un error interno durante el procesado de los datos |

## Obtener identificador de expediente

Recibe:

| Campo | Descripcion | Tamaño |
| ----- | ------------| -------|
| **unidad** | código dir3 de la unidad productora | 9 dígitos |
|  **anio** | el año para el documento |4 dígitos |
|  **serie** | información de la clasificación lógica, serie y subserie documental |8 dígitos |

Resultado:

| Estado HTTP | Descripción |
| ----------- | ----------- |
| 200 | JSON con el campo **identificador** del siguiente identificador de expediente para la unidad y año |
| 400 | si alguno de los parámetros no tiene los valores correctos |
| 500 | si se produce un error interno durante el procesado de los datos |


## Base de datos

Se hace uso del SGBD postgresql, al arrancar crea, si no existen, las tablas y resto de elementos necesarios para la ejecución del proyecto

En concreto se crea:

- Tabla: goeniidentificador que almacena para cada expediente o documento, unidad y año el siguiente número de identificador a generar.

## Variables de entorno

El proyecto usa dos variables de entorno:

- DATABASE_CONN_STRING con la cadena de conexión a la base de datos postgresql
- DATABASE_SCHEMA esquema de la base de datos en que se encuentran las tablas

