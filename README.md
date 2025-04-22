# RAGScholar

RAGScholar is a research paper search and analysis platform that leverages modern LLMs (Gemini), vector search (Qdrant), and scalable microservices with concurrency via RabbitMQ. The system enables fast, context-aware search and explanation for research papers.

---


## Video



https://github.com/user-attachments/assets/4713cef6-59e9-47d4-a821-c952de3421b7



## Screenshots

<img width="1452" alt="Screenshot 2025-04-22 at 11 02 16 PM" src="https://github.com/user-attachments/assets/7f730227-070d-42d4-a95e-953a73248929" />

<img width="1452" alt="Screenshot 2025-04-22 at 11 02 14 PM" src="https://github.com/user-attachments/assets/e9a73085-af31-4e1f-ba3d-73c8479ab791" />

<img width="1452" alt="Screenshot 2025-04-22 at 11 01 41 PM" src="https://github.com/user-attachments/assets/7b53758c-2baf-45a2-8dbb-1ff0a566e5c2" />

<img width="1452" alt="Screenshot 2025-04-22 at 11 01 33 PM" src="https://github.com/user-attachments/assets/f7c18e8b-5c38-4d73-a9cb-0589e03a3311" />



## Architecture Overview

<img width="1154" alt="Screenshot 2025-04-22 at 11 29 24 PM" src="https://github.com/user-attachments/assets/8466b8bc-2445-4a23-9bd6-92c805a74202" />


- **Frontend (client/):** Next.js app with TypeScript, Tailwind CSS, and shadcn/ui. Provides search, paper detail, and context-aware search UI.
- **Service (server/service):** Main backend API. Handles search requests, fetches from Arxiv DB, and coordinates Gemini LLM for explanations. Publishes tasks to RabbitMQ for concurrent processing.
- **Consumer (server/consumer):** Worker service that consumes tasks from RabbitMQ, runs Gemini LLM and vector DB queries, and returns results.
- **RabbitMQ:** Message queue for distributing concurrent tasks between Service and Consumer.
- **Qdrant Vector DB:** Stores vector embeddings for semantic search and retrieval.
- **Arxiv DB:** Source of research paper metadata/content (fetched in parallel).

---

## Components

### 1. `client/` (Frontend)
- Next.js app for user interaction.
- Search bar for global and context-aware search.
- Calls Service API for search and explanations.

### 2. `server/service/` (Service)
- Go service exposing the main API (e.g., `/analyze`).
- Handles:
  - Receiving search/explanation requests from frontend
  - Fetching papers from Arxiv DB (parallel fetching)
  - Publishing tasks to RabbitMQ for concurrent Gemini/vector processing
  - Aggregating results

### 3. `server/consumer/` (Consumer)
- Go service acting as a worker/consumer for RabbitMQ.
- Handles:
  - Consuming tasks (search/explanation)
  - Running Gemini LLM for explanations
  - Querying Qdrant for semantic search
  - Returning results to Service

### 4. `RabbitMQ`
- Message broker for decoupling and concurrency between Service and Consumer.
- Allows multiple consumers for scalability.

### 5. `Qdrant Vector DB`
- Vector database for storing and querying paper embeddings.
- Used for fast semantic search and retrieval.

### 6. `Arxiv DB`
- Source of research paper metadata and full text.
- Fetched in parallel by Service.

---

## Running Locally

1. **Start Infrastructure:**
   - From `server/`, run:
     ```bash
     docker-compose up -d
     ```
   - This starts RabbitMQ and Qdrant (see `server/docker-compose.yml`).

2. **Start Backend:**
   - In `server/service/` and `server/consumer/`, build and run each Go service:
     ```bash
     go run main.go
     ```

3. **Start Frontend:**
   - In `client/`, install dependencies and run:
     ```bash
     npm install
     npm run dev
     ```
   - Visit [http://localhost:3000](http://localhost:3000)

---

## Notes
- The system is designed for high concurrency and scalability.
- Modify `.env` files for custom RabbitMQ/Qdrant endpoints if needed.
- See the architecture diagram (`docs/architecture.png`) for data flow and component interaction.

---

## Credits
- Built with Next.js, Go, RabbitMQ, Qdrant, and Google Gemini.
