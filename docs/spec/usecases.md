# Use Cases

## UC01 — Chat Interativo

**Ator principal**: Usuário (desenvolvedor via OpenCode/Superpowers)

**Fluxo principal**:
1. Usuário envia prompt via chat
2. Engine seleciona o modelo carregado ou o default
3. Engine monta o contexto com histórico da sessão
4. Provider executa inferência no dispositivo selecionado
5. Engine retorna resposta token por token (streaming) ou completa
6. Engine atualiza histórico da sessão

**Fluxo alternativo — Sem sessão ativa**:
1. Engine cria nova sessão
2. Continua do passo 2 do fluxo principal

**Fluxo de exceção — Modelo não carregado**:
1. Engine retorna erro 503 (modelo não disponível)
2. Cliente deve carregar modelo antes de tentar novamente

---

## UC02 — Autocomplete (Code Completion)

**Ator principal**: VS Code (via extensão OpenCode)

**Fluxo principal**:
1. VS Code envia contexto do código (prefixo + sufixo)
2. Engine prepara prompt especializado para completion
3. Provider executa inferência no modelo de code completion
4. Engine retorna sugestão de continuação
5. VS Code exibe sugestão inline

---

## UC03 — Embedding de Texto

**Ator principal**: Sistema (RAG pipeline, skill, ou plugin)

**Fluxo principal**:
1. Sistema envia texto para endpoint /embedding
2. Provider carrega modelo BGE Small (se não estiver carregado)
3. Engine verifica cache de embedding
4. Se cache hit: retorna embedding armazenado
5. Se cache miss: executa inferência, armazena em cache, retorna

---

## UC04 — Reranking de Documentos

**Ator principal**: Sistema (RAG pipeline)

**Fluxo principal**:
1. Sistema envia query + lista de documentos
2. Provider executa modelo de reranking
3. Engine retorna documentos reordenados por relevância

---

## UC05 — Gerenciamento de Modelos

**Ator principal**: Usuário ou Sistema

**Fluxo — Load**:
1. Usuário envia requisição POST /model/load com model_id
2. Provider verifica se modelo já está carregado
3. Se não: carrega modelo do disco para o dispositivo alvo
4. Retorna status do carregamento

**Fluxo — Unload**:
1. Usuário envia requisição POST /model/unload com model_id
2. Provider descarrega modelo da memória
3. Retorna confirmação

**Fluxo — Listagem**:
1. Usuário envia GET /models
2. Engine varre diretório de modelos
3. Retorna lista com status de cada modelo

---

## UC06 — Detecção de Dispositivos

**Ator principal**: Sistema (inicialização) ou Usuário

**Fluxo principal**:
1. Sistema (ou usuário) solicita GET /devices
2. Provider detecta CPU, GPU, NPU disponíveis
3. Sistema retorna lista com capacidades de cada dispositivo
4. Engine seleciona automaticamente o melhor dispositivo disponível

---

## UC07 — Benchmark de Modelo

**Ator principal**: Usuário (desenvolvedor)

**Fluxo principal**:
1. Usuário solicita benchmark para modelo específico
2. Engine executa bateria de testes (latência, throughput, memória)
3. Engine retorna relatório detalhado
4. Relatório pode ser exportado para comparação

---

## UC08 — Execução de Skill

**Ator principal**: Usuário ou Sistema

**Fluxo principal**:
1. Usuário solicita execução de skill por nome
2. Engine carrega definição da skill (YAML/JSON)
3. Engine executa pipeline de passos da skill
4. Cada passo pode ser: prompt, embedding, rerank, condicional
5. Engine retorna resultado final da skill
