import { Command } from "../types"

export default {
  name: 'ytmp3',
  alias: ['ytaudio', 'yta'],
  category: 'downloader',
  description: 'Download YouTube audio (MP3)',
  wait: true,
  async handler(bot, args, context) {
    if (!args.length) {
      return bot.sendMessage(context.chat,
        '⚠️ Please provide a YouTube URL\nExample: /ytmp3 https://youtu.be/dQw4w9WgXcQ'
      )
    }

    const url = args[0]
    const youtubeRegex = /(youtu\.be\/|youtube\.com\/(watch\?v=|embed\/|v\/|shorts\/))([a-zA-Z0-9_-]{11})/
    if (!youtubeRegex.test(url)) {
      return bot.sendMessage(context.chat,
        '❌ Invalid YouTube URL. Please provide a valid link.'
      )
    }

    try {
      const result = await bot.downloader(url, 'youtube', 'mp3')
      console.log('YouTube MP3 Result:', result)

      if (!result?.result?.url) {
        throw new Error(result?.error || 'No audio URL received')
      }

      await bot.sendAudio(context.chat, result.result.url, true)
      await bot.sendMessage(context.chat, 
        `✅ Download complete!\nTitle: ${result.result.title || 'Unknown'}\nDuration: ${result.result.duration ? `${Math.floor(result.result.duration / 60)}m ${result.result.duration % 60}s` : 'Unknown'}`
      )
    } catch (error: any) {
      console.error('[YouTube MP3] Download error:', error)
      await bot.sendMessage(
        context.chat, 
        `❌ Failed to download YouTube audio: ${error.message}`
      )
    }
  }
} as Command