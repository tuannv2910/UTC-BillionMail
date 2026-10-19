import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockPost = vi.fn()

vi.mock('@/api', () => ({
	instance: {
		post: (...args: any[]) => mockPost(...args),
	},
}))

import { aiGenerateContent, aiReviewContent, getAiModels } from './ai'

describe('ai API', () => {
	beforeEach(() => {
		vi.clearAllMocks()
	})

	it('getAiModels calls POST /askai/supplier/models', () => {
		getAiModels()
		expect(mockPost).toHaveBeenCalledWith('/askai/supplier/models')
	})

	it('aiReviewContent calls POST /email_template/ai_review', () => {
		const params = {
			subject: 'Hi',
			content: '<p>x</p>',
			language: 'en',
			unsubscribe_managed: true,
			supplier_name: 'Gemini',
			model_id: 'm',
		}
		aiReviewContent(params)
		expect(mockPost).toHaveBeenCalledWith('/email_template/ai_review', params)
	})

	it('aiGenerateContent calls POST /email_template/ai_generate', () => {
		const params = {
			brief: 'Launch',
			subject: '',
			current_html: '',
			tone: 'friendly',
			language: 'vi',
			supplier_name: '',
			model_id: '',
		}
		aiGenerateContent(params)
		expect(mockPost).toHaveBeenCalledWith('/email_template/ai_generate', params)
	})
})
