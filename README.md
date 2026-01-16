# API Sistema de Incidencias

Este es el backend para el Sistema de Incidencias, desarrollado en Go con el framework Gin.

## Requisitos Previos

Antes de ejecutar el proyecto, asegúrate de tener instalado:

1.  **Go** (Golang): [Descargar e instalar](https://go.dev/dl/)
2.  **PostgreSQL**: [Descargar e instalar](https://www.postgresql.org/download/)

## Configuración de Base de Datos

El proyecto utiliza una configuración de base de datos específica (hardcoded en `database/database.go`). Para que conecte correctamente, debes configurar tu PostgreSQL local de la siguiente manera:

1.  **Base de datos**: `lc_db`
2.  **Usuario**: `myuser`
3.  **Contraseña**: `test1234`
4.  **Puerto**: `5432` (Puerto por defecto)

Puedes crear el usuario y la base de datos ejecutando los siguientes comandos en tu terminal de SQL (`psql` o pgAdmin):

```sql
CREATE USER myuser WITH PASSWORD 'test1234';
CREATE DATABASE lc_db OWNER myuser;
ALTER USER myuser WITH SUPERUSER; -- Opcional, si necesitas permisos elevados
```

> **Nota:** Si prefieres usar tus propias credenciales, debes editar el archivo `database/database.go` y modificar la línea `dsn`:
> ```go
> dsn := "host=127.0.0.1 user=TU_USUARIO password=TU_CONTRASEÑA dbname=TU_DB port=5432 sslmode=disable"
> ```

## Instalación

1.  Clona el repositorio (si no lo has hecho ya):
    ```bash
    git clone <url-del-repo>
    cd apiSistemaIncidencias
    ```

2.  Instala las dependencias del proyecto:
    ```bash
    go mod tidy
    ```

## Ejecución

Para iniciar el servidor, ejecuta el siguiente comando en la raíz del proyecto:

```bash
go run main.go
```

El servidor iniciará en el puerto **8000**.
Deberías ver un mensaje como:
```
Conexión a la base de datos exitosa.
Migración de la base de datos exitosa.
[GIN-debug] Listening and serving HTTP on :8000
```

## Frontend

El backend está configurado para aceptar peticiones (CORS) desde `http://localhost:5173`.
Esto indica que el frontend debe ejecutarse en ese puerto para comunicarse correctamente con esta API.

Clona el repositorio del frontend, asegúrate de iniciarlo con `npm run dev` y que esté corriendo en el puerto 5173.
