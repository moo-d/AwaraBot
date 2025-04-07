<div align="center">
  <img src="https://files.catbox.moe/ugs4hj.jpg" width="150" alt="AwaraBot Logo">
  <h1>AwaraBot</h1>
  <p>WhatsApp Bot (Go + TypeScript)</p>
  
  [![Go Version](https://img.shields.io/badge/Go-1.20%2B-blue?logo=go)](https://golang.org/)
  [![Node Version](https://img.shields.io/badge/Node-18%2B-green?logo=node.js)](https://nodejs.org/)
  [![License](https://img.shields.io/badge/License-MIT-red)](LICENSE)
  [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/moo-d/AwaraBot/pulls)
</div>

## 🌟 Key Features

### 🎵 Media Downloader
| Service  | Format           | Example Command       |
|----------|------------------|-----------------------|
| YouTube  | MP3 Audio        | `/yta https://youtu.be/...` |
| YouTube  | MP4 Video (720p) | `/ytv https://youtu.be/...` |
| TikTok   | Video (No WM)    | `/tt https://vm.tiktok.com/...` |

### 🛠️ Utility Commands
```bash
/menu   # Show all commands
/stats  # Show bot statistics
```

## 🚀 Getting Started

### ⚙️ Installation
```bash
# 1. Clone repository
git clone https://github.com/moo-d/AwaraBot && cd AwaraBot

# 2. Install dependencies
go mod tidy && npm install

# 3. Configure environment
cp .env.example .env
nano .env  # Edit your configuration

# 4. Start the bot
npm start
```

### 🔧 Configuration
```env
# .env Example
BOT_NAME=Awara
```

## 🖥️ Tech Stack

| Component       | Technology               |
|-----------------|--------------------------|
| Backend         | Go (whatsmeow library)   |
| Command System  | TypeScript               |
| IPC             | STDIN/STDOUT             |
| Database        | SQLite (embedded)        |
|Media Processing	| FFmpeg                   |

---

## 📢 Join For More Information
<a href="https://chat.whatsapp.com/L1xOwYMceo64Ff8958Q1rT"> <img src="https://img.shields.io/badge/Join_Group-25D366?style=for-the-badge&logo=whatsapp&logoColor=white" alt="WhatsApp Group"> </a> </div>
