export function extractCommand(input: string): {
  prefix: string | null
  command: string
  args: string[]
} {
  const trimmed = input.trim()
  if (!trimmed) return { prefix: null, command: '', args: [] }

  const [cmdPart, ...rawArgs] = trimmed.split(/\s+/)
  const args = rawArgs.filter(Boolean)

  const hasPrefix = /^[\/!\.#~]/.test(cmdPart)
  const prefix = hasPrefix ? cmdPart[0] : null
  const command = hasPrefix ? cmdPart.slice(1) : cmdPart

  return { prefix, command, args }
}
