# Instalación y ejecución

## Ejecución rápida con Docker Compose

1) Levanta los servicios:
   ```bash
   docker compose up
   ```
2) La API quedará disponible en:
   - `http://localhost:8080/healthz`
   - `http://localhost:8080/v1/docs`
   - `POST http://localhost:8080/v1/migrate`
   - `GET http://localhost:8080/v1/users/{user_id}/balance?from=YYYY-MM-DDThh:mm:ssZ&to=YYYY-MM-DDThh:mm:ssZ`

Se recominda usar `http://localhost:8080/v1/docs` para realizar las pruebas desde la implementación con swagger

## Desarrollo con Dev Container

Requisitos:
- Docker y Docker Compose
- Visual Studio Code
- Extensión “Dev Containers”

Pasos:
1) Instala la extensión “Dev Containers” en VS Code.
2) Abre este repositorio en VS Code.
3) Opción A: Click en la esquina inferior izquierda (><) y selecciona “Rebuild Container...”.  
   Opción B: Abre la barra de comandos (macOS: Cmd+Shift+P, Windows/Linux: Ctrl+Shift+P) y ejecuta “Dev Containers: Rebuild Container...”.

