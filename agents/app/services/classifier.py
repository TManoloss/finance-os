import json
import asyncio
import re
from app.services.llm_factory import get_llm_provider
from app.models.transaction import TransactionClassifyRequest

# Regras estáticas para estabelecimentos comuns (Economiza IA)
STATIC_RULES = {
	# Alimentação
	r"IFOOD|IFD |RAPPI|ZE DELIVERY|RESTAURANTE|LANCHONETE|PADARIA|PANIFICADORA|DOCERIA|SORVETERIA|CAFE|COFFEE|BURGER|MCDONALDS|BK |SPOLETO|OAKBERRY|CHURRASCARIA|PIZZARIA|HAMBURGUERIA|GARCIA": "Alimentação",
	r"SUPERMERCADO|ATACADAO|CARREFOUR|PAO DE ACUCAR|CONDOR|MUFFATO|PARANA SUPERMERCADO|COOP |MERCADO|HORTIFRUTI|SACOLAO|CONVENIENCIA": "Alimentação",
	
	# Transporte
	r"UBER|99APP|CABIFY|ESTACIONAMENTO|ESTAC |PEDAGIO|SEM PARAR|VELOE|CONCESSIONARIA|LOCALIZA": "Transporte",
	r"POSTO|GASOLINA|COMBUSTIVEL|SHELL|IPIRANGA|PETROBRAS|ALCOOL|ETANOL|AUTO POSTO": "Transporte",
	
	# Assinaturas
	r"NETFLIX|SPOTIFY|APPLE.COM|GOOGLE.*STORAGE|ICLOUD|DISNEYPLUS|HBO|AMAZON PRIME|KINDLE|CRUNCHYROLL|YOUTUBE|CANVA": "Assinaturas",
	
	# Saúde
	r"DROGASIL|RAIA|PACHECO|FARMACIA|FARM |CLINICA|LABORATORIO|HOSPITAL|MEDICO|DENTISTA|ODONTO|EXAME|UNIMED|DROGARIA": "Saúde",
	r"ESPACOLASER": "Saúde",

	# Lazer
	r"CINEMA|CINE |SHOW|VIAGEM|HOTEL|AIRBNB|DECOLAR|BOOKING|TURISMO|HOBBY|JOGO|STEAM|PLAYSTATION|XBOX|NINTENDO": "Lazer",
	r"SHOPEE|AMAZON|MERCADO.?LIVRE|MAGAZINE.*LUIZA|CASAS.*BAHIA|ALIEXPRESS|SHEIN": "Lazer",
	
	# Moradia
	r"ALUGUEL|CONDOMINIO|COPEL|SANEPAR|ENEL|SABESP|LUZ|AGUA|INTERNET|VIVO|CLARO|TIM|OI |CONTA DE LUZ|CONTA DE AGUA": "Moradia",
	
	# Pet
	r"PETZ|COBASI|VETERINARIO|PET SHOP|ANIMAL|RAÇÃO": "Pet",
	
	# Educação
	r"CURSO|LIVRO|MENSALIDADE|COLEGIO|ESCOLA|FACULDADE|UNIVERSIDADE|UDEMY|HOTMART|ALURA": "Educação",
	
	# Investimentos
	r"INVESTIMENTOS|CDB|TESOURO|APORT|XP INVESTIMENTOS|BTG PACTUAL|CORRETORA|NUINVEST|INTER INVEST": "Investimentos",
	
	# Outros (Padrões que sabemos que são outros ou internos)
	r"IOF|JUROS|TARIF|TARIFA|MANUTENCAO CONTA|PAGAMENTO ON LINE|PAG BOLETO|PAGTO DEBITO": "Outros",
    
    # Renda
	r"SALARIO|PROVENTO|DIVIDENDOS|RENDIMENTO|TRANSFERENCIA RECEBIDA|PIX RECEBIDO|DOC RECEBIDO|TED RECEBIDO|RESGATE|CORRECAO MONETARIA|TRANSFERÊNCIA RECEBIDA": "Renda",

	# Transferência (Último recurso para transações bancárias puras)
	r"PIX ENVIADO|TRANSFERENCIA ENVIADA|TRANSFERÊNCIA ENVIADA|DOC ENVIADO|TED ENVIADO|PAGAMENTO FATURA": "Transferência",
}

SYSTEM_PROMPT = """
Você é um especialista em classificação de transações financeiras para o mercado brasileiro.
Sua tarefa é receber os dados de uma transação e retornar o nome da categoria mais adequada.

ATENÇÃO: No Brasil, muitas transações Pix/Transferência são gastos reais. Se a descrição contiver um nome de empresa (ex: GARCIA HAMBURGUERIA, UBER, IFOOD), ignore o fato de ser uma "Transferência" e classifique pelo estabelecimento.

Categorias disponíveis:
- Alimentação (restaurantes, delivery, supermercados, padarias)
- Transporte (uber, 99, combustível, estacionamento, pedágio)
- Saúde (farmácias, hospitais, consultas, exames)
- Lazer (cinema, shows, viagens, hobbies, compras de variedades, shopee, amazon)
- Assinaturas (netflix, spotify, saas, icloud, streaming)
- Moradia (aluguel, condomínio, luz, água, internet, materiais de construção)
- Educação (cursos, livros, mensalidades, escolas)
- Investimentos (corretoras, aportes, bolsa)
- Renda (salário, dividendos, vendas, pix recebido de clientes)
- Pet (petshop, veterinário, ração)
- Emergências (imprevistos, multas)
- Transferência (APENAS para transferências entre contas do mesmo titular ou entre pessoas físicas sem objetivo de compra)
- Outros (quando não se encaixar em nenhuma anterior)

Regras:
1. Responda APENAS um JSON válido no formato: {"category_id": "NOME_EXATO_DA_CATEGORIA", "confidence": 0.0}
2. Priorize o nome do estabelecimento sobre o método de pagamento (Pix/TED).
3. Se houver dúvidas, use Outros.
"""

class ClassifierService:
	def __init__(self):
		self.llm = get_llm_provider()
		self._cache = {} # Cache simples em memória

	async def classify(self, tx: TransactionClassifyRequest) -> dict:
		# Limpeza de prefixos comuns de Pix/Transferência
		def clean_text(text: str) -> str:
			if not text: return ""
			# Remove "Transferência enviada|", "Pix enviado|", etc.
			text = re.sub(r"^(TRANSFER[ÊE]NCIA|PIX|TED|DOC)\s+(ENVIAD[OA]|RECEBID[OA])\s*[|:]\s*", "", text, flags=re.IGNORECASE)
			# Remove CPFs/CNPJs e IDs longos
			text = re.sub(r"\d{3}\.\d{3}\.\d{3}-\d{2}|\d{2}\.\d{3}\.\d{3}/\d{4}-\d{2}", "", text)
			return text.strip().upper()

		desc_clean = clean_text(tx.description)
		merchant_clean = clean_text(tx.merchant_name)
		combined = f"{merchant_clean} {desc_clean}".strip()
		
		# Se após limpar o texto ficou vazio ou manteve apenas o original, tenta usar o original sem prefixos
		if not combined:
			combined = f"{tx.merchant_name} {tx.description}".upper()

		print(f"[Classifier] Analisando: '{combined}' (Original: {tx.description})")

		# 1. Regras Estáticas (Regex)
		for pattern, category in STATIC_RULES.items():
			if re.search(pattern, combined):
				# Se for um padrão de transferência genérico, só aceita se não tivermos nada melhor
				if category == "Transferência" and any(re.search(p, combined) for p in list(STATIC_RULES.keys())[:-1]):
					continue
					
				print(f"[Classifier] REGRA: '{pattern}' MATCHED '{combined}' -> {category}")
				return {"category_id": category, "confidence": 1.0}

		# 2. Verificar Cache
		cache_key = f"{tx.merchant_name}|{tx.description}|{tx.direction}".lower()
		if cache_key in self._cache:
			return self._cache[cache_key]

		# 3. IA (Caso as regras não peguem)
		prompt = f"""
		Classifique esta transação:
		Estabelecimento: {tx.merchant_name}
		Descrição: {tx.description}
		Valor: R$ {tx.amount}
		Direção: {tx.direction}
		"""

		try:
			# Pequeno atraso para evitar rate limit massivo no sync
			print(f"[Classifier] Throttling 4.5s para: {tx.description}")
			await asyncio.sleep(4.5) 
			
			try:
				response_text = await self.llm.completion(prompt, system_prompt=SYSTEM_PROMPT)
			except Exception as e:
				if "429" in str(e):
					print(f"[Classifier] Rate limit atingido. Aguardando 10s para retry...")
					await asyncio.sleep(10.0)
					response_text = await self.llm.completion(prompt, system_prompt=SYSTEM_PROMPT)
				else:
					raise e
			
			# Limpa possíveis markdown da resposta do LLM
			clean_json = response_text.replace("```json", "").replace("```", "").strip()
			if not clean_json:
				return {"category_id": "Outros", "confidence": 0.0}

			result = json.loads(clean_json)
			final_result = {
				"category_id": result.get("category_id", "Outros"),
				"confidence": result.get("confidence", 0.0)
			}
			
			# Salva no cache
			self._cache[cache_key] = final_result
			return final_result

		except Exception as e:
			print(f"Erro na classificação via IA para '{tx.description}': {e}")
			return {"category_id": "Outros", "confidence": 0.0}
