# 1. Detener los contenedores de Podman

podman compose down

# 2. Detener PostgreSQL local

sudo systemctl stop postgresql-17 / puede ser ejecutado desde cualquer ubicación

# 3. Verificar que todo esté detenido

podman ps
sudo systemctl status postgresql-17 --no-pager // puede ser ejecutado desde cualquer ubicación

# Opción A — con contenedores (recomendado)

podman compose up -d // ejecutar en la carpeta del proyecto

# Opción B — desarrollo local

sudo systemctl start postgresql-17// puede ser ejecutado desde cualquer ubicación
make run

# task-manager
