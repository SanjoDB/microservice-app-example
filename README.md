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

### 3. Estrategias de Branching

#### 3.1. Estrategia de Branching para Desarrollo (GitHub Flow)
Rama principal: master

**Flujo de trabajo:**

- Se crea una rama __feature/nombre-funcionalidad__ para cada nueva funcionalidad, corrección de error o mejora.

- Se realiza un pull request (PR) desde la rama feature hacia master.

**Code Review:**

Todo PR debe ser revisado antes de ser fusionado. De esta tarea se encarga tanto los desarrolladores como los Pipelines definidos en el proyecto.

Al aprobarse el PR, se pasaran los cambios de la rama feature a master y se realizara el despliegue del codigo tras su integración.

Ejemplo de ramas feature:

- feature/Retry-Pattern

- feature/Circuit-Breaker


#### 3.2. Estrategia de Branching para Operaciones (Rama por Entorno)
Ramas principales:

- infra/dev

- infra/prod

**Flujo de trabajo:**

Cada rama gestiona la infraestructura de un entorno específico (dev y prod).

Las variables sensibles (por ejemplo, IPs públicas, nombres de recursos, credenciales mínimas) están definidas separadamente por entorno.

Cada rama contiene scripts de infraestructura declarativa en Terraform.

**Comportamiento:**

- **infra/dev:** crea recursos para ambiente de desarrollo. En este ambiente se despliega el codigo de las ramas **Feature** y sirve para probar los cambios que se realizan mientras se desarrolla el codigo.

- **infra/prod:** crea recursos para ambiente de producción. En este ambiente se despliega el codigo de la rama **Master** y sirve para mostrar el producto final, es la ejecicion de codigo ya aprovado.

---

### 4. Patrones Cloud Implementados

#### 4.1. Patrón Retry

### Descripción
El patrón Retry permite reintentar operaciones que han fallado, especialmente útil para operaciones transitorias como llamadas de red o acceso a bases de datos.

### Implementación en los Microservicios
- Auth API (Go)

```go
func Retry[T any](config RetryConfig, operation func() (T, error)) (T, error) {
    var result T
    var err error
    waitTime := config.WaitTime

    for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
        result, err = operation()
        if err == nil {
            return result, nil
        }
        if waitTime < config.MaxWaitTime {
            waitTime *= 2
        }
        time.Sleep(waitTime)
    }
    return result, err
}
```

- TODOs API (Node.js)
```javascript
async function retry(config, operation) {
    let waitTime = config.waitTime;
    for (let attempt = 1; attempt <= config.maxAttempts; attempt++) {
        try {
            return await operation();
        } catch (error) {
            if (attempt === config.maxAttempts) break;
            if (waitTime < config.maxWaitTime) waitTime *= 2;
            await new Promise(resolve => setTimeout(resolve, waitTime));
        }
    }
}
```

- Log Message Processor (Python)
```python
@retry(retry_config)
def connect_redis(host, port):
    def _connect():
        return redis.Redis(host=host, port=port, db=0)
    return redis_cb.execute(_connect)
```

### Beneficios

- Manejo automático de fallos temporales
- Backoff exponencial para evitar sobrecarga
- Mejora la resiliencia del sistema
- Configuración flexible (intentos, tiempos de espera)

#### 4.2. Patrón Circuit Breaker

### Descripción
El Circuit Breaker previene que una aplicación continúe ejecutando operaciones que probablemente fallarán, permitiendo que el sistema se recupere y evitando fallos en cascada.

### Implementación en los Microservicios

- Auth API (Go)
```go
type CircuitBreaker[T any] struct {
    mutex           sync.RWMutex
    state           State
    failureCount    int
    failureThreshold int
    resetTimeout    time.Duration
    lastFailureTime time.Time
}
```

- TODOs API (Node.js)
```javascript
class CircuitBreaker {
    constructor(failureThreshold = 3, resetTimeout = 10000) {
        this.state = States.CLOSED;
        this.failureCount = 0;
        this.failureThreshold = failureThreshold;
        this.resetTimeout = resetTimeout;
    }
}
```

- Log Message Processor (Python)
```Python
class CircuitBreaker:
    def __init__(self, failure_threshold=3, reset_timeout=10):
        self._state = State.CLOSED
        self._failure_count = 0
        self._failure_threshold = failure_threshold
        self._reset_timeout = reset_timeout
```

### Beneficios

- Prevención de fallos en cascada
- Recuperación gradual del sistema
- Protección de recursos
- Mejor manejo de errores sistémicos
- Monitorización del estado del sistema

#### 4.3. Patrón External Configuration

### Descripción
External Configuration permite externalizar la configuración de la aplicación fuera del código, permitiendo cambios sin necesidad de recompilar.

### Implementación

- Azure Pipelines
```YML
- task: CopyFilesOverSSH@0
    inputs:
        sshEndpoint: VM-DEV-CONNECTION
        sourceFolder: $(service_dir)
        contents: '**'
        targetFolder: $(remote_service_dir)
    displayName: 'Copiar carpeta actualizada a la VM'
```

```YML
  - task: AzureCLI@2
    inputs:
      azureSubscription: 'Kevin'
      scriptType: 'bash'
      scriptLocation: 'inlineScript'
      inlineScript: |
        export ARM_CLIENT_ID=$(ARM_CLIENT_ID)
        export ARM_CLIENT_SECRET=$(ARM_CLIENT_SECRET)
        export ARM_SUBSCRIPTION_ID=$(ARM_SUBSCRIPTION_ID)
        export ARM_TENANT_ID=$(ARM_TENANT_ID)
```

### Beneficios

- Configuración sin recompilación
- Diferentes configuraciones por ambiente
- Gestión centralizada de configuraciones
- Mayor seguridad al externalizar secretos
- Facilita CI/CD y despliegues

---

### 5. Analisis y funcionamiento de las Pipelines de Azure DevOps

#### Estructura general de las Pipelines
1. **Pipelines de Infraestructura (Por Entorno)**

- azure-pipeline-DEV.yml
- azure-pipeline-PROD.yml

2. **Pipelines de Desarrollo (Por Microservicio)**

- auth-api-dev.yml
- users-api-dev.yml
- todos-api-dev.yml
- log-message-processor-dev.yml
- frontend-dev.yml

3. **Pipelines de Control**

- pullRequest-Validation.yml - Validación de PRs
- deploy-Production.yml - Despliegue a producción

#### 5.1 Analisis de Pipelines

#### 5.1.1. Pipelines de Infraestructura

**Caracteristica común**
```YML
trigger:
  branches:
    include:
      - infra/Entorno
```
- **Trigger:** Se activan en cambios a rama infra/Entorno
- **Ambiente:** Crea la VM de desarrollo (IP: 172.172.175.246) o Crea la VM de produccion (IP: 20.172.180.248)

- **Pasos Comunes:**
1. Instala Terraform
2. Verifica la version de Terraform
3. Establece Python para AzureCLI
4. Se conecta a AzureCLI
5. Ejecuta Terraform init
6. Ejecuta Terraform Plan
7. Ejecuta Terraform Apply

#### 5.1.2. Pipelines de Desarrollo

**Caracteristica común**
```YML
trigger:
  branches:
    include:
      - feature/*
      - master
  paths:
    include:
      - service-name/**
```

- **Trigger:** Se activan en cambios a rama master o ramas feature/*

- **Path Filtering:** Solo se ejecutan si hay cambios en su carpeta específica

- **Ambiente:** Usan la VM de desarrollo (IP: 172.172.175.246)

- **Pasos Comunes:**
1. Checkout del código
2. Copia de archivos vía SSH
3. Build con Docker Compose
4. Despliegue del servicio
5. Validación del estado del contenedor

#### 5.1.3 Pipelines de Control

**Validacion de PR**
```YML
trigger: none
pr:
  branches:
    include:
      - master
      - prod
```

- **Propósito:** Validar que todo compile antes de permitir merge

- **Trigger:** Solo en Pull Requests

- **Validaciones:**
1. Compila todos los servicios
2. Verifica Dockerfiles
3. Verifica dependencias

**Pipeline de Producción**
```YML
trigger:
  branches:
    include:
      - master
pr: none
```

- **Propósito:** Despliegue a producción

- **Trigger:** Solo en cambios a master

- **Ambiente:** VM de producción (IP: 20.172.180.248)

- **Proceso:**
1. Copia completa del repositorio
2. Build de todos los servicios
3. Despliegue completo
4. Validación de servicios críticos

---

### 5. Flujo del Sistema

1. El usuario hace login desde el frontend → auth-api.

2. Se genera un JWT y se usa para acceder a todos-api.

3. Cada operación genera un log enviado a Redis.

4. log-message-processor consume mensajes de Redis y los registra.

5. users-api ofrece información de usuario solicitada por auth-api.

---

### 6. Diagrama de Arquitectura

![microservice-app-example](/arch-img/Diagrama.png)
