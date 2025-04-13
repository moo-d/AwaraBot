export function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  const secs = Math.floor(seconds % 60)

  const parts = []
  if (days > 0) parts.push(`${days} Hari`)
  parts.push(`${hours} Jam`, `${mins} Menit`, `${secs} Detik`)
  
  return parts.join(', ')
}

export async function getCpuUsage(): Promise<string> {
  return new Promise((resolve) => {
    const start = process.cpuUsage()
    setTimeout(() => {
      const end = process.cpuUsage(start)
      const usage = (end.user + end.system) / 10000
      resolve(usage.toFixed(2) + '%')
    }, 100)
  })
}