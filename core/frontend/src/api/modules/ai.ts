import { instance } from '@/api'
import type { AiGenerateParams, AiReviewParams } from '@/features/AiContent/types'

/**
 * @description Enabled AI models configured under Settings > AI Model
 */
export const getAiModels = () => {
	return instance.post('/askai/supplier/models')
}

export const aiReviewContent = (params: AiReviewParams) => {
	return instance.post('/email_template/ai_review', params)
}

export const aiGenerateContent = (params: AiGenerateParams) => {
	return instance.post('/email_template/ai_generate', params)
}
