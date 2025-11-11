# Stori Challenge API

API para migración de transacciones vía CSV. 
Expone documentación OpenAPI y una ruta de salud.

Consulta la instalación y ejecución en [docs/installation.md](docs/installation.md).

## Sobre el proyecto

La API contiene los siguientes endpoints principales:

### `POST /v1/migrate`

Se encarga de migrar transacciones desde archivos CSV hacia una base de datos PostgreSQL.

#### Restricciones y reglas de validación:
- Si el ID de transacción viene duplicado dentro del archivo, se marca error.
- Si el ID ya existe en la base de datos, se marca error.
- Si algún dato no cumple con el formato esperado, se marca error.
- El tamaño máximo permitido del archivo es 5 MB, pensado para migraciones rápidas.

#### Flujo básico:
1. El endpoint recibe el archivo CSV.
2. Valida estructura, duplicados y datos existentes.
3. Inserta los registros válidos en una transacción atómica (rollback ante cualquier error).
4. Devuelve un resumen del resultado.

### `GET /v1/users/{user_id}/balance`

Devuelve el balance de un usuario y los totales de débitos y créditos dentro de un rango de tiempo opcional.

- Parámetros:
  - `user_id` (path): ID del usuario.
  - `from` (query, opcional): Fecha/hora inferior en formato RFC3339 con `Z`.
  - `to` (query, opcional): Fecha/hora superior en formato RFC3339 con `Z`.
    - Si solo viene uno de los dos, ese se usa como límite inferior y el superior es “ahora (UTC)”.
    - Si vienen ambos, se ordenan automáticamente; el mayor es el límite superior.
    - El límite superior no puede ser mayor que “ahora (UTC)”.
- Respuesta 200:
```json
{
  "balance": 25.21,
  "total_debits": 10,
  "total_credits": 15
}
```
- Respuesta 404: si el `user_id` no tiene ninguna transacción registrada.
- Respuesta 400: si `from` o `to` no cumplen el formato.

## Mejoras futuras con más tiempo

El endpoint actual de `/v1/migrate` no utiliza goroutines ni worker pools, 
ya que está pensado para ejecutarse dentro de una AWS Lambda. 
Por cuestiones de tiempo, no se subió la Lambda, pero la API está diseñada para migraciones rápidas y pequeñas (≤5 MB).

Para archivos más pesados, se planea la creación de un nuevo endpoint:

### `POST /v1/migrate-async`

#### Objetivo:
Procesar migraciones grandes de manera asíncrona, delegando el procesamiento pesado a un worker.

#### Flujo propuesto:
1. El endpoint no recibe un archivo directamente, sino una URL donde el CSV ya se encuentra almacenado (por ejemplo en S3).  
   Esto evita consumir recursos de la Lambda solo para subir archivos grandes (operación costosa).
2. Crea un registro en una nueva tabla `migrations`, con los siguientes campos:
   - `id`  
   - `url`  
   - `status` (ej. `PENDING`, `PROCESSING`, `COMPLETED`, `FAILED`)
3. Lanza un evento SQS que será procesado por un worker (según el tamaño del dataset):
   - Si el proceso dura <15 min → AWS Lambda  
   - Si requiere más tiempo → EC2, Fargate o AWS Glue (con Spark o Apache Beam)

#### Consideraciones de diseño:
- **Idempotencia:**  
  El servicio será idempotente, garantizando que ejecutar la misma solicitud más de una vez no genere duplicados ni inconsistencias.
- **Control de concurrencia:**  
  Antes de procesar, el worker revisará el `status` en la tabla `migrations`.  
  Si está en `PROCESSING`, significa que otro proceso ya lo tomó.
- **Evitar reenvíos duplicados de SQS:**  
  Se configurará un `VisibilityTimeout` alto y se usará `ChangeMessageVisibility` periódicamente para extender el tiempo de procesamiento sin que el mensaje sea reenviado.

### Otras mejoras planificadas

- **Factories de testing:**  
  Implementar factories para simplificar la creación de datos en tests de integración.
  
- **Optimización de balance:**  
  Actualmente los balances se calculan a partir de todas las transacciones.  
  En escenarios con gran volumen, convendría mantener una tabla `balances` que se actualice incrementalmente con cada transacción, reduciendo el costo computacional de los cálculos.

- **CI/CD e Infraestructura:**  
  Implementar un workflow de GitHub Actions para despliegue automático en AWS (con Infrastructure as Code, IaC).  
  Esto permitirá tener pipelines reproducibles, control de versiones de infraestructura y despliegues automatizados a distintos entornos.

- **Optimización de usuarios**  
  Crear una tabla `users` para facilitar la búsqueda y filtrado por usuario y rango de tiempo.  
  Esto permitiría optimizar consultas relacionales entre usuarios y transacciones, especialmente en escenarios de alto volumen.  
  Sin embargo, dado que el reto menciona exclusivamente la migración de registros de transacciones, se asumió que la tabla `users` ya existe en el sistema.

## Cosas extra implementadas

Además de la funcionalidad principal, se realizaron varias mejoras técnicas adicionales:

- **Tests unitarios e integraciones:**  
  Se implementaron pruebas unitarias y de integración para garantizar la estabilidad del sistema y la correcta interacción con la base de datos.

- **Soporte OpenAPI 3.1 con Swagger:**  
  La documentación de la API se genera dinámicamente en formato OpenAPI 3.1, compatible con Swagger UI, permitiendo probar los endpoints directamente desde el navegador.

- **Ejecución optimizada de migraciones en tests:**  
  Las migraciones de base de datos solo se ejecutan una vez por proceso de pruebas, utilizando `sync.Once`.  
  Además, los tests de integración se ejecutan sobre conexiones controladas por `txdb`, donde cada test se ejecuta dentro de una transacción aislada que se revierte automáticamente al finalizar (`db.Close()`).  
  Esto asegura que la base de datos siempre regrese a su estado original, manteniendo un entorno limpio y consistente para cada prueba.

- **Respuestas detalladas de errores por fila en `/v1/migrate`:**  
  El endpoint de migración reporta errores a nivel de fila en los archivos CSV, indicando exactamente en qué línea ocurrió el problema.  
  Esto facilita la depuración por parte del usuario y mejora la experiencia de carga de datos.
