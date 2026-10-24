/**
 * Encodings tried, in order, when a text file is not valid UTF-8.
 * CSV files saved by Excel are often GB18030/GBK (Chinese Windows) or
 * Windows-1252 (Western Windows) instead of UTF-8, which showed up as garbled
 * attributes after import (#153, #212).
 */
const FALLBACK_ENCODINGS = ['gb18030', 'windows-1252'] as const

const stripBOM = (text: string) => (text.charCodeAt(0) === 0xfeff ? text.slice(1) : text)

const tryDecode = (bytes: Uint8Array, encoding: string, fatal: boolean): string | null => {
	try {
		return new TextDecoder(encoding, { fatal }).decode(bytes)
	} catch {
		return null
	}
}

/**
 * Decode file bytes into text: UTF-8 (with or without BOM) first, then common
 * legacy encodings. The BOM is always stripped so the first CSV header
 * ("email") is recognised.
 */
export const decodeTextBytes = (input: ArrayBuffer | Uint8Array): string => {
	const bytes = input instanceof Uint8Array ? input : new Uint8Array(input)

	const utf8 = tryDecode(bytes, 'utf-8', true)
	if (utf8 !== null) return stripBOM(utf8)

	for (const encoding of FALLBACK_ENCODINGS) {
		const text = tryDecode(bytes, encoding, true)
		if (text !== null) return stripBOM(text)
	}

	// Last resort: lossy UTF-8 (same as FileReader.readAsText)
	return stripBOM(new TextDecoder('utf-8').decode(bytes))
}

/**
 * Read a File/Blob as text with encoding detection.
 */
export const readTextFile = async (file: Blob): Promise<string> => {
	const buffer = await file.arrayBuffer()
	return decodeTextBytes(buffer)
}
