## Documentación y Entrega

### 1. Objetivo del Proyecto
Desarrollar e implementar una arquitectura de microservicios que permita gestionar tareas (todos), autenticación de usuarios y procesamiento de logs, integrando prácticas modernas de DevOps (CI/CD en Azure DevOps) y patrones de arquitectura cloud como **Health Endpoint Monitoring** y **Gateway Aggregation**.

---

### 2. Estructura de Microservicios

| Microservicio           | Lenguaje         | Descripción                                                                 |
|-------------------------|------------------|-----------------------------------------------------------------------------|
| `auth-api`              | Go               | Maneja autenticación y generación de tokens JWT.                            |
| `users-api`             | Java (Spring)    | Exposición de usuarios y verificación de credenciales.                      |
| `todos-api`             | Node.js          | Gestión de tareas: creación, listado y eliminación.                         |
| `log-message-processor` | Python           | Procesa logs desde una cola Redis.                                          |
| `frontend`              | Vue/JS           | Interfaz gráfica que interactúa con los demás microservicios.               |

---

### 3. CI/CD con Azure DevOps

- Pipelines separados para `build` y `deploy` de cada microservicio.
- En entorno **Dev**, el despliegue se realiza en una VM Linux usando **PM2**.
- En **Staging** y **Prod**, las VMs realizan `git pull` y compilan localmente tras recibir PR.

---

### 4. Patrones Cloud Implementados

#### Health Endpoint Monitoring
Todos los microservicios exponen un endpoint HTTP `/health` que retorna:
```json
{ "status": "UP" }
```

#### Gateway Aggregation
El frontend actúa como consumidor de múltiples servicios:

auth-api: login y JWT

users-api: validación de usuario

todos-api: operaciones CRUD

Se planea agregar un gateway unificado en etapas posteriores.

### 5. Flujo del Sistema

El usuario hace login desde el frontend → auth-api.

Se genera un JWT y se usa para acceder a todos-api.

Cada operación genera un log enviado a Redis.

log-message-processor consume mensajes de Redis y los registra.

users-api ofrece información de usuario solicitada por auth-api.

### 6. Diagrama de Arquitectura

Mirar la carpeta docs

### 7. Recursos Técnicos

Variables de entorno detalladas en cada microservicio.

Todos los servicios responden en /health.

Pruebas funcionales realizadas con curl y validadas en VM Dev.

ZIPKIN utilizado para trazabilidad distribuida (en todos los servicios).

### 8. Evidencias de Entrega

Repositorio GitHub: microservice-app-example

Pipelines en Azure DevOps corriendo correctamente.

Diagrama de arquitectura actualizado.

Logs de ejecución incluidos en cada VM por medio de pm2 logs.