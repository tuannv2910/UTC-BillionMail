import { describe, it, expect } from 'vitest'
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, resolve } from 'node:path'
import en from './lang/en.json'
import zh from './lang/zh.json'
import ja from './lang/ja.json'
import vi from './lang/vi.json'

type Messages = Record<string, unknown>

const flatten = (obj: unknown, prefix = '', out: Record<string, string> = {}) => {
	if (obj && typeof obj === 'object' && !Array.isArray(obj)) {
		for (const [k, v] of Object.entries(obj as Messages)) flatten(v, prefix ? `${prefix}.${k}` : k, out)
	} else {
		out[prefix] = String(obj)
	}
	return out
}

const placeholders = (s: string) => (s.match(/\{[A-Za-z0-9_]+\}/g) || []).sort().join(',')

const enFlat = flatten(en)

const srcRoot = resolve(__dirname, '..')
const sourceFiles = (dir: string): string[] =>
	readdirSync(dir).flatMap(name => {
		const p = join(dir, name)
		if (statSync(p).isDirectory()) return sourceFiles(p)
		return /\.(vue|ts|tsx)$/.test(name) && !/\.(test|d)\.ts$/.test(name) ? [p] : []
	})

describe('locale files', () => {
	for (const [lang, messages] of Object.entries({ zh, ja, vi })) {
		it(`${lang} keeps the placeholders of en`, () => {
			const flat = flatten(messages)
			const broken = Object.keys(enFlat).filter(
				k => k in flat && placeholders(flat[k]) !== placeholders(enFlat[k])
			)
			expect(broken).toEqual([])
		})
	}

	it('vi has no empty translations', () => {
		const empty = Object.entries(flatten(vi)).filter(([k, v]) => !v.trim() && enFlat[k]?.trim())
		expect(empty).toEqual([])
	})

	it('every editor/ui key used in the source exists in en', () => {
		const used = new Set<string>()
		for (const file of sourceFiles(srcRoot)) {
			const content = readFileSync(file, 'utf8')
			for (const m of content.matchAll(/['"`]((?:editor|ui)\.[A-Za-z0-9_.]+)['"`]/g)) used.add(m[1])
		}
		const missing = [...used].filter(k => !(k in enFlat) && !k.startsWith('editor.blocks.'))
		expect(missing).toEqual([])
	})
})
