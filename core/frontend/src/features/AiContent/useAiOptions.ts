/**
 * @description Languages the AI can write in
 */
export const useAiLanguageOptions = () => {
	const { t } = useI18n()
	return computed(() => [
		{ label: t('template.aiContent.language.en'), value: 'en' },
		{ label: t('template.aiContent.language.vi'), value: 'vi' },
		{ label: t('template.aiContent.language.zh'), value: 'zh' },
		{ label: t('template.aiContent.language.ja'), value: 'ja' },
	])
}

/**
 * @description Writing tones; the value is sent to the model as is
 */
export const useAiToneOptions = () => {
	const { t } = useI18n()
	return computed(() => [
		{ label: t('template.aiContent.tone.professional'), value: 'professional' },
		{ label: t('template.aiContent.tone.friendly'), value: 'friendly and warm' },
		{ label: t('template.aiContent.tone.persuasive'), value: 'persuasive but honest' },
		{ label: t('template.aiContent.tone.concise'), value: 'short and concise' },
	])
}
