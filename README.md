# Awara Bot Hybrid (Go + TypeScript)
 
*A hybrid WhatsApp bot with Go backend and TypeScript command system*

<!--## 🔥 Features -->

## 🛠️ Tech Stack

| Component       | Technology               |
|-----------------|--------------------------|
| Backend         | Go (whatsmeow library)   |
| Command System  | TypeScript               |
| IPC             | STDIN/STDOUT             |
| Database        | SQLite (embedded)        |

## 🚀 Installation

### Prerequisites
- Go 1.20+
- Node.js 1.18+
- WhatsApp mobile app

### Setup
1. Clone the repository
   ```bash
   git clone https://github.com/moo-d/AwaraBot
   cd whatsapp-bot
   ```

2. Install dependencies
   ```bash
   go mod download
   npm install
   ```

3. Configure environment
   ```bash
   cp .env.example .env
   nano .env  # Edit with your config
   ```

4. Build and run
   ```bash
   chmod +x build.sh
   ./build.sh && npm run prod
   ```
