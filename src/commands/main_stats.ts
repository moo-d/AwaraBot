import os from 'os'
import process from 'process'
import { formatUptime, getCpuUsage } from '../utils/statsHelper'
import { Command } from '../types'

export default {
  name: 'stats',
  alias: ['stat'],
  category: 'main',
  description: 'Check bot statistics',
  async handler(bot, args, context) {
    console.log(context.isGroup)
    const startTime = process.hrtime()
    
    const memoryUsage = process.memoryUsage()
    const cpuInfo = os.cpus()[0] || {}
    
    const stats = {
      responseSpeed: 0,
      uptimeBot: formatUptime(process.uptime()),
      uptimeServer: formatUptime(os.uptime()),
      memoryUsage: (memoryUsage.rss / 1048576).toFixed(2), // 1024*1024
      cpuModel: cpuInfo.model || 'Unknown',
      cpuSpeed: cpuInfo.speed || 0,
      cpuUsage: await getCpuUsage(),
      platform: os.platform(),
      arch: os.arch(),
      ramTotal: (os.totalmem() / 1073741824).toFixed(2), // 1024^3
      ramFree: (os.freemem() / 1073741824).toFixed(2)
    }

    const diff = process.hrtime(startTime)
    stats.responseSpeed = parseFloat((diff[0] * 1000 + diff[1] / 1e6).toFixed(2))

    const message = `
╭━━━〔 📊 BOT STATISTICS 〕━━━╮
│
│  🔹 Bot Status:
│  ├ 🚀 Response Speed: ${stats.responseSpeed} ms
│  ├ ⏳ Uptime Bot: ${stats.uptimeBot}
│  ├ ⏳ Uptime Server: ${stats.uptimeServer}
│  ├ 📂 Memory Usage: ${stats.memoryUsage} MB
│  
│  🖥 Server Info:
│  ├ 🔧 CPU Model: ${stats.cpuModel}
│  ├ ⚡ CPU Speed: ${stats.cpuSpeed} MHz
│  ├ 📊 CPU Usage: ${stats.cpuUsage}%
│  
│  📜 Additional Info:
│  ├ 🌐 Platform: ${stats.platform}
│  ├ 🏷 Arch: ${stats.arch}
│  ├ 💾 RAM Total: ${stats.ramTotal} GB
│  ├ 📉 RAM Free: ${stats.ramFree} GB
│
╰━━━━━━━━━━━━━━━━━━━╯`.trim()

    await bot.sendMessage(context.chat, message)
  }
} as Command