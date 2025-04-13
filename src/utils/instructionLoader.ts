import fs from 'fs';
import path from 'path';

const INSTRUCTIONS_DIR = path.join(__dirname, '../../instructions');

export function loadInstruction(filename: string): { role: 'user'; content: string } {
  try {
    const content = fs.readFileSync(path.join(INSTRUCTIONS_DIR, filename), 'utf-8');
    return { role: 'user' as const, content };
  } catch (error) {
    console.error(`Error loading instruction ${filename}:`, error);
    return { role: 'user', content: '' };
  }
}

export const BASE_INSTRUCTIONS = [
  'personality.txt',
  'capabilities.txt',
  'response_format.txt'
].map(loadInstruction).filter(inst => inst.content);