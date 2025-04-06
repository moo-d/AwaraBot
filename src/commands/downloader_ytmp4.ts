import { Command } from '../types'

export default {
  name: 'ytmp4',
  alias: ['ytvideo', 'ytv'],
  category: 'downloader',
  description: 'Download YouTube video (MP4)',
  async handler(bot, args, context) {
    if (!args.length) {
      return bot.sendMessage(
        context.chat,
        '⚠️ Please provide a YouTube URL\nExample: /ytmp4 https://youtu.be/dQw4w9WgXcQ'
      )
    }

    const url = args[0]
    const youtubeRegex = /(youtu\.be\/|youtube\.com\/(watch\?v=|embed\/|v\/|shorts\/))([a-zA-Z0-9_-]{11})/
    
    if (!youtubeRegex.test(url)) {
      return bot.sendMessage(
        context.chat,
        '❌ Invalid YouTube URL. Please provide a valid YouTube link.'
      )
    }

    try {
      await bot.sendMessage(context.chat, '⏳ Downloading YouTube video... (This may take a while)')

      const { result, error } = await bot.downloader(url, 'youtube', 'mp4')

      if (!result?.url) {
        throw new Error(error || 'No video URL received')
      }

      if (!result.url.startsWith('http')) {
        throw new Error('Invalid video URL format')
      }

      const [minutes, seconds] = result.duration 
        ? [Math.floor(result.duration / 60), Math.floor(result.duration % 60)]
        : [0, 0]

      await Promise.all([
        bot.sendVideo(context.chat, result.url, result.title || 'YouTube Video'),
        bot.sendMessage(
          context.chat,
          `✅ *${result.title || 'YouTube Video'}*\n` +
          (result.duration ? `⏱ Duration: ${minutes}m ${seconds}s` : '')
        )
      ])

    } catch (error: any) {
      await bot.sendMessage(
        context.chat,
        `❌ Failed to download video: ${error.message || 'Unknown error'}\n` +
        'Please try again later.'
      )
    }
  }
} as Command