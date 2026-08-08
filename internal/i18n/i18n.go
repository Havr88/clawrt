package i18n

import (
	"fmt"
	"strings"
)

type Lang string

const (
	ES Lang = "es"
	EN Lang = "en"
	FR Lang = "fr"
	PT Lang = "pt"
	IT Lang = "it"
	RU Lang = "ru"
	ZH Lang = "zh"
	JA Lang = "ja"
	AR Lang = "ar"
)

type Catalog map[string]map[Lang]string

var dict = Catalog{
	"help_title": {
		ES: "🐾 *ClawRT — Agente IA para OpenWrt*",
		EN: "🐾 *ClawRT — OpenWrt AI Agent*",
		FR: "🐾 *ClawRT — Agent IA pour OpenWrt*",
		PT: "🐾 *ClawRT — Agente IA para OpenWrt*",
		IT: "🐾 *ClawRT — Agente IA per OpenWrt*",
		RU: "🐾 *ClawRT — ИИ-агент для OpenWrt*",
		ZH: "🐾 *ClawRT — OpenWrt AI 智能代理*",
		JA: "🐾 *ClawRT — OpenWrt AIエージェント*",
		AR: "🐾 *ClawRT — وكيل الذكاء الاصطناعي لـ OpenWrt*",
	},
	"help_body": {
		ES: "Comandos rápidos disponibles:\n• `/status` o `/sysinfo` - Estado del router y recursos\n• `/wifi` - Estado de la red WiFi\n• `/reboot` - Reiniciar el router\n• `/help` - Muestra esta ayuda\n\n💡 *Conversación abierta:* Escríbeme cualquier consulta y usaré las herramientas para gestionar tu router.",
		EN: "Available quick commands:\n• `/status` or `/sysinfo` - Router & resource status\n• `/wifi` - WiFi network status\n• `/reboot` - Reboot router\n• `/help` - Show this help\n\n💡 *Open Chat:* Ask me anything and I will use ClawRT tools to manage your router.",
		FR: "Commandes rapides disponibles :\n• `/status` ou `/sysinfo` - État du routeur et ressources\n• `/wifi` - État du réseau WiFi\n• `/reboot` - Redémarrer le routeur\n• `/help` - Afficher cette aide\n\n💡 *Chat ouvert :* Posez-moi vos questions e j'utiliserai les outils de ClawRT.",
		PT: "Comandos rápidos disponíveis:\n• `/status` ou `/sysinfo` - Status do roteador e recursos\n• `/wifi` - Status da rede WiFi\n• `/reboot` - Reiniciar roteador\n• `/help` - Mostrar ajuda\n\n💡 *Chat aberto:* Faça qualquer pergunta e usarei as ferramentas para gerenciar seu roteador.",
		IT: "Comandi rapidi disponibili:\n• `/status` o `/sysinfo` - Stato del router e risorse\n• `/wifi` - Stato della rete WiFi\n• `/reboot` - Riavvia router\n• `/help` - Mostra aiuto\n\n💡 *Chat aperta:* Fammi qualsiasi domanda e userò gli strumenti per gestire il tuo router.",
		RU: "Доступные команды:\n• `/status` или `/sysinfo` - Состояние роутера и ресурсов\n• `/wifi` - Состояние WiFi\n• `/reboot` - Перезагрузка роутера\n• `/help` - Справка\n\n💡 *Чат:* Задайте любой вопрос, и я использую инструменты ClawRT для управления роутером.",
		ZH: "可用快捷命令：\n• `/status` 或 `/sysinfo` - 路由器及资源状态\n• `/wifi` - WiFi 网络状态\n• `/reboot` - 重启路由器\n• `/help` - 显示帮助\n\n💡 *自由对话：* 随时提问，我将使用 ClawRT 工具管理您的路由器。",
		JA: "利用可能なコマンド：\n• `/status` または `/sysinfo` - ルーターとリソースの状態\n• `/wifi` - WiFiネットワークの状態\n• `/reboot` - ルーター再起動\n• `/help` - ヘルプを表示\n\n💡 *オープンチャット：* ご質問があればClawRTツールを使用して管理します。",
		AR: "الأوامر السريعة المتاحة:\n• `/status` أو `/sysinfo` - حالة الموجه والموارد\n• `/wifi` - حالة شبكة WiFi\n• `/reboot` - إعادة تشغيل الموجه\n• `/help` - عرض المساعدة\n\n💡 *محادثة مفتوحة:* اسألني أي شيء وسأستخدم أدوات ClawRT لإدارة الموجه الخاص بك.",
	},
	"status_header": {
		ES: "📊 *Estado de %s*\n• *Sistema:* %s (%s)\n• *Uptime:* %s\n• *Carga CPU:* %s\n• *Memoria RAM:* %d MB / %d MB (%.1f%% uso)\n\n%s",
		EN: "📊 *Status of %s*\n• *System:* %s (%s)\n• *Uptime:* %s\n• *CPU Load:* %s\n• *RAM Memory:* %d MB / %d MB (%.1f%% used)\n\n%s",
		FR: "📊 *État de %s*\n• *Système :* %s (%s)\n• *Temps d'activité :* %s\n• *Charge CPU :* %s\n• *Mémoire RAM :* %d MB / %d MB (%.1f%% utilisé)\n\n%s",
		PT: "📊 *Status de %s*\n• *Sistema:* %s (%s)\n• *Uptime:* %s\n• *Carga CPU:* %s\n• *Memória RAM:* %d MB / %d MB (%.1f%% uso)\n\n%s",
		IT: "📊 *Stato di %s*\n• *Sistema:* %s (%s)\n• *Uptime:* %s\n• *Carico CPU:* %s\n• *Memoria RAM:* %d MB / %d MB (%.1f%% uso)\n\n%s",
		RU: "📊 *Статус %s*\n• *Система:* %s (%s)\n• *Время работы:* %s\n• *Загрузка ЦП:* %s\n• *Память ОЗУ:* %d МБ / %d МБ (%.1f%% исп.)\n\n%s",
		ZH: "📊 *%s 的状态*\n• *系统：* %s (%s)\n• *运行时间：* %s\n• *CPU 负载：* %s\n• *内存：* %d MB / %d MB (已用 %.1f%%)\n\n%s",
		JA: "📊 *%s のステータス*\n• *システム：* %s (%s)\n• *稼働時間：* %s\n• *CPU負荷：* %s\n• *RAMメモリ：* %d MB / %d MB (%.1f%% 使用中)\n\n%s",
		AR: "📊 *حالة %s*\n• *النظام:* %s (%s)\n• *وقت التشغيل:* %s\n• *حمولة المعالج:* %s\n• *الذاكرة:* %d ميجابايت / %d ميجابايت (%.1f%% مستخدمة)\n\n%s",
	},
	"executing_tool": {
		ES: "🛠️ *Ejecutando herramienta:* `%s`...",
		EN: "🛠️ *Executing tool:* `%s`...",
		FR: "🛠️ *Exécution de l'outil :* `%s`...",
		PT: "🛠️ *Executando ferramenta:* `%s`...",
		IT: "🛠️ *Esecuzione dello strumento:* `%s`...",
		RU: "🛠️ *Выполнение инструмента:* `%s`...",
		ZH: "🛠️ *正在执行工具：* `%s`...",
		JA: "🛠️ *ツールを実行中：* `%s`...",
		AR: "🛠️ *جاري تنفيذ الأداة:* `%s`...",
	},
	"processing_llm": {
		ES: "🧠 *Procesando solicitud con ClawRT LLM...*",
		EN: "🧠 *Processing request with ClawRT LLM...*",
		FR: "🧠 *Traitement de la demande avec ClawRT LLM...*",
		PT: "🧠 *Processando solicitação com ClawRT LLM...*",
		IT: "🧠 *Elaborazione della richiesta con ClawRT LLM...*",
		RU: "🧠 *Обработка запроса с помощью ClawRT LLM...*",
		ZH: "🧠 *正在使用 ClawRT LLM 处理请求...*",
		JA: "🧠 *ClawRT LLMでリクエストを処理中...*",
		AR: "🧠 *جاري معالجة الطلب بواسطة ClawRT LLM...*",
	},
	"sys_prompt": {
		ES: "Eres ClawRT, un agente autónomo de IA experto en OpenWrt y administración de routers domésticos. Tienes herramientas para consultar el estado del router, interfaces de red, leer y modificar configuraciones UCI y ejecutar diagnósticos. Sé conciso y preciso en idioma español.",
		EN: "You are ClawRT, an autonomous AI agent expert in OpenWrt and home router administration. You have tools to check router status, network interfaces, read/modify UCI configs, and run diagnostics. Be concise and precise in English.",
		FR: "Vous êtes ClawRT, un agent IA autonome expert d'OpenWrt et d'administration de routeurs. Soyez précis en français.",
		PT: "Você é o ClawRT, um agente autônomo de IA especialista em OpenWrt e administração de roteadores. Seja preciso em português.",
		IT: "Sei ClawRT, un agente IA autonomo esperto di OpenWrt e amministrazione di router. Sii preciso in italiano.",
		RU: "Вы ClawRT, автономный ИИ-агент, эксперт по OpenWrt и администрированию роутеров. Будьте точны на русском языке.",
		ZH: "你是 ClawRT，一个精通 OpenWrt 和路由器管理的自主 AI 代理。请用简明准确的中文回答。",
		JA: "あなたはClawRT、OpenWrtおよびルーター管理の専門AIエージェントです。日本語で簡洁かつ正確に回答してください。",
		AR: "أنت ClawRT، وكيل ذكاء اصطناعي مستقل خبير في OpenWrt وإدارة الموجهات. كن موجزاً ودقيقاً باللغة العربية.",
	},
}

func NormalizeLang(code string) Lang {
	code = strings.ToLower(strings.TrimSpace(code))
	if strings.HasPrefix(code, "es") { return ES }
	if strings.HasPrefix(code, "en") { return EN }
	if strings.HasPrefix(code, "fr") { return FR }
	if strings.HasPrefix(code, "pt") { return PT }
	if strings.HasPrefix(code, "it") { return IT }
	if strings.HasPrefix(code, "ru") { return RU }
	if strings.HasPrefix(code, "zh") { return ZH }
	if strings.HasPrefix(code, "ja") { return JA }
	if strings.HasPrefix(code, "ar") { return AR }
	return ES // Default Spanish
}

func T(lang Lang, key string, args ...interface{}) string {
	l := NormalizeLang(string(lang))
	if catalog, ok := dict[key]; ok {
		if text, found := catalog[l]; found {
			if len(args) > 0 {
				return fmt.Sprintf(text, args...)
			}
			return text
		}
		// Fallback to English if translation missing for specific key
		if text, found := catalog[EN]; found {
			if len(args) > 0 {
				return fmt.Sprintf(text, args...)
			}
			return text
		}
	}
	return key
}
