import { describe, expect, it } from 'vitest'
import { decodeTextBytes, readTextFile } from './file'

describe('decodeTextBytes', () => {
	it('decodes UTF-8 and strips the BOM', () => {
		const bytes = new Uint8Array([0xef, 0xbb, 0xbf, ...new TextEncoder().encode('email,attributes')])
		expect(decodeTextBytes(bytes)).toBe('email,attributes')
	})

	it('keeps UTF-8 Vietnamese text intact', () => {
		const text = 'a@x.com,"{""name"":""Nguyễn Văn A""}"'
		expect(decodeTextBytes(new TextEncoder().encode(text))).toBe(text)
	})

	it('falls back to GB18030 for Chinese Excel CSV files', () => {
		// "姓名" encoded as GBK
		const gbk = new Uint8Array([0xd0, 0xd5, 0xc3, 0xfb])
		expect(decodeTextBytes(gbk)).toBe('姓名')
	})

	it('reads a Blob', async () => {
		const blob = new Blob([new Uint8Array([0xef, 0xbb, 0xbf, 0x68, 0x69])])
		await expect(readTextFile(blob)).resolves.toBe('hi')
	})
})
