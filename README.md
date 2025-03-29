![Captura de pantalla](ss.png)

# La Liga Tracker ⚽

Proyecto web con frontend, backend y base de datos en contenedores Docker, para registrar y gestionar partidos de fútbol.

---

## 🐳 Requisitos

- Docker
- Docker Compose

---

## 🛠️ Configuración

Ya se proporcionan los archivos `.env.example` y `docker-compose.example.yml`.

1. Crea los archivos `.env` y `docker-compose.yml` a partir de los ejemplos:
   ```bash
   cp .env.example .env
   cp docker-compose.example docker-compose.yml
   ```

2. Revisa y ajusta las variables si es necesario.

---

## 🚀 Ejecución del proyecto

1. Construye y levanta los contenedores:
   ```bash
   docker compose up --build
   ```

2. Abre el navegador en las siguientes URLs:
   - **Frontend:** [http://localhost:3000](http://localhost:3000)
   - **Backend:** [http://localhost:8080/api/matches](http://localhost:8080/api/matches)

---

## 📬 Documentación de Postman

👉 [Ver colección de Postman](https://www.postman.com/restless-crescent-464318/my-workspace/collection/kws0hu9/api-partidos?action=share&source=copy-link&creator=40770150)

---

## 🧾 Licencia

MIT. Hecho con pasión, goles y un poco de sudor digital.

