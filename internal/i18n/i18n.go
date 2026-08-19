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
		ES: "Comandos rápidos disponibles:\n• `/status` o `/sysinfo` - Estado del router, CPU, RAM y red\n• `/clients` o `/dhcp` - Lista de clientes conectados a la LAN\n• `/wifi` - Estado de la red inalámbrica WiFi\n• `/qrwifi` o `/qr` - Código QR de acceso a WiFi\n• `/scan` - Escáner de puertos inseguros en la LAN\n• `/models` - Consultar modelos LLM disponibles (Bynara/Provider)\n• `/clear` - Vaciar hechos aprendidos en memoria\n• `/reboot` - Reiniciar servicio ClawRT\n• `/help` - Muestra esta ayuda\n\n💡 *Conversación abierta:* Escríbeme cualquier consulta y usaré las herramientas para gestionar tu router.",
		EN: "Available quick commands:\n• `/status` or `/sysinfo` - Router, CPU, RAM & network status\n• `/clients` or `/dhcp` - List of connected LAN devices\n• `/wifi` - WiFi wireless network status\n• `/qrwifi` or `/qr` - WiFi access QR code\n• `/scan` - LAN security port scanner\n• `/models` - Query active LLM models (Bynara/Provider)\n• `/clear` - Clear learned facts memory\n• `/reboot` - Reboot ClawRT service\n• `/help` - Show this help\n\n💡 *Open Chat:* Ask me anything and I will use ClawRT tools to manage your router.",
		FR: "Commandes rapides disponibles :\n• `/status` ou `/sysinfo` - État du routeur et ressources\n• `/clients` ou `/dhcp` - Liste des appareils connectés\n• `/wifi` - État du réseau WiFi\n• `/qrwifi` ou `/qr` - Code QR d'accès WiFi\n• `/scan` - Scanner de ports réseau\n• `/models` - Modèles LLM disponibles\n• `/clear` - Effacer la mémoire\n• `/reboot` - Redémarrer le service\n• `/help` - Afficher cette aide\n\n💡 *Chat ouvert :* Posez-moi vos questions et j'utiliserai les outils de ClawRT.",
		PT: "Comandos rápidos disponíveis:\n• `/status` ou `/sysinfo` - Status do roteador, CPU e RAM\n• `/clients` ou `/dhcp` - Dispositivos conectados na LAN\n• `/wifi` - Status da rede WiFi\n• `/qrwifi` ou `/qr` - Código QR do WiFi\n• `/scan` - Escáner de portas inseguras na LAN\n• `/models` - Modelos LLM disponíveis\n• `/clear` - Limpar memória\n• `/reboot` - Reiniciar serviço\n• `/help` - Mostrar ajuda\n\n💡 *Chat aberto:* Faça qualquer pergunta e usarei as ferramentas para gerenciar seu roteador.",
		IT: "Comandi rapidi disponibili:\n• `/status` o `/sysinfo` - Stato del router e risorse\n• `/clients` o `/dhcp` - Dispositivi connessi in LAN\n• `/wifi` - Stato della rete WiFi\n• `/qrwifi` o `/qr` - Codice QR per WiFi\n• `/scan` - Scanner di porte di sicurezza\n• `/models` - Modelli LLM disponibili\n• `/clear` - Cancella memoria aprendida\n• `/reboot` - Riavvia servizio\n• `/help` - Mostra aiuto\n\n💡 *Chat aperta:* Fammi qualsiasi domanda e userò gli strumenti per gestire il tuo router.",
		RU: "Доступные команды:\n• `/status` или `/sysinfo` - Состояние роутера, ЦП и ОЗУ\n• `/clients` или `/dhcp` - Подключенные устройства LAN\n• `/wifi` - Состояние WiFi\n• `/qrwifi` или `/qr` - QR-код для WiFi\n• `/scan` - Сканирование портов LAN\n• `/models` - Доступные модели ИИ\n• `/clear` - Очистить память\n• `/reboot` - Перезапуск службы\n• `/help` - Справка\n\n💡 *Чат:* Задайте любой вопрос, и я использую инструменты ClawRT для управления роутером.",
		ZH: "可用快捷命令：\n• `/status` 或 `/sysinfo` - 路由器、CPU、内存及网络状态\n• `/clients` 或 `/dhcp` - 已连接的 LAN 设备列表\n• `/wifi` - WiFi 无线网络状态\n• `/qrwifi` 或 `/qr` - WiFi 连接二维码\n• `/scan` - LAN 局域网端口扫描\n• `/models` - 查询可用 AI 模型\n• `/clear` - 清空已学记忆\n• `/reboot` - 重启 ClawRT 服务\n• `/help` - 显示帮助\n\n💡 *自由对话：* 随时提问，我将使用 ClawRT 工具管理您的路由器。",
		JA: "利用可能なコマンド：\n• `/status` または `/sysinfo` - ルーターとリソースの状態\n• `/clients` または `/dhcp` - 接続されているLANデバイス一覧\n• `/wifi` - WiFiネットワークの状態\n• `/qrwifi` または `/qr` - WiFi接続用QRコード\n• `/scan` - LANセキュリティポートスキャン\n• `/models` - 利用可能なAIモデル一覧\n• `/clear` - 学習メモリを消去\n• `/reboot` - ClawRTサービス再起動\n• `/help` - ヘルプを表示\n\n💡 *オープンチャット：* ご質問があればClawRTツールを使用して管理します。",
		AR: "الأوامر السريعة المتاحة:\n• `/status` أو `/sysinfo` - حالة الموجه والمعالج والذاكرة\n• `/clients` أو `/dhcp` - قائمة الأجهزة المتصلة بالشبكة\n• `/wifi` - حالة شبكة WiFi\n• `/qrwifi` أو `/qr` - رمز QR للاتصال بـ WiFi\n• `/scan` - فحص المنافذ في الشبكة\n• `/models` - نماذج الذكاء الاصطناعي المتاحة\n• `/clear` - مسح الذاكرة\n• `/reboot` - إعادة تشغيل الخدمة\n• `/help` - عرض المساعدة\n\n💡 *محادثة مفتوحة:* اسألني أي شيء وسأستخدم أدوات ClawRT لإدارة الموجه الخاص بك.",
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
	if strings.HasPrefix(code, "es") {
		return ES
	}
	if strings.HasPrefix(code, "en") {
		return EN
	}
	if strings.HasPrefix(code, "fr") {
		return FR
	}
	if strings.HasPrefix(code, "pt") {
		return PT
	}
	if strings.HasPrefix(code, "it") {
		return IT
	}
	if strings.HasPrefix(code, "ru") {
		return RU
	}
	if strings.HasPrefix(code, "zh") {
		return ZH
	}
	if strings.HasPrefix(code, "ja") {
		return JA
	}
	if strings.HasPrefix(code, "ar") {
		return AR
	}
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
		if text, found := catalog[EN]; found {
			if len(args) > 0 {
				return fmt.Sprintf(text, args...)
			}
			return text
		}
	}
	return key
}
