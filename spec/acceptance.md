# Acceptance Criteria

## Model Management

| # | Critério | Como Verificar |
|---|----------|----------------|
| AC01 | Modelo OpenVINO deve carregar corretamente | POST /model/load retorna status "loaded" |
| AC02 | Modelo deve descarregar da memória | POST /model/unload retorna status "unloaded" |
| AC03 | Modelo já carregado não deve recarregar | POST /model/load duas vezes retorna 200, não 500 |
| AC04 | Lista de modelos deve refletir diretório | GET /models retorna todos os modelos disponíveis |
| AC05 | Modelo inválido deve retornar erro 404 | POST /model/load com ID inexistente retorna 404 |
| AC06 | Troca dinâmica de modelo deve funcionar | Chat em modelo A → troca para B → chat usa B |

## Device Detection

| # | Critério | Como Verificar |
|---|----------|----------------|
| AC07 | CPU deve ser detectada | GET /devices mostra CPU disponível |
| AC08 | GPU Intel deve ser detectada (se presente) | GET /devices mostra GPU se hardware disponível |
| AC09 | NPU Intel deve ser detectada (se presente) | GET /devices mostra NPU se hardware disponível |
| AC10 | Dispositivo automático deve selecionar GPU > NPU > CPU | Engine escolhe GPU se disponível, senão NPU, senão CPU |
| AC11 | Dispositivo manual deve ser respeitado | Inferência com device="CPU" usa CPU |

## Inference

| # | Critério | Como Verificar |
|---|----------|----------------|
| AC12 | Chat inference deve retornar texto coerente | POST /chat com prompt retorna resposta relevante |
| AC13 | Completion deve completar texto | POST /completion com prefixo retorna continuação |
| AC14 | Streaming deve retornar chunks | POST /chat com stream=true retorna SSE events |
| AC15 | Parâmetros temperature/top_k/top_p devem afetar output | Mesmo prompt com params diferentes → outputs diferentes |
| AC16 | Max tokens deve limitar resposta | Resposta não excede max_tokens |
| AC17 | Time to first token < 500ms (após carga) | Medido no campo timing.ttft |

## Embeddings

| # | Critério | Como Verificar |
|---|----------|----------------|
| AC18 | Embedding deve retornar vetor de floats | POST /embeddings retorna array de float32 |
| AC19 | Embedding de textos similares deve ter alta similaridade | Similaridade coseno > 0.9 para textos equivalentes |
| AC20 | Embedding de textos diferentes deve ter baixa similaridade | Similaridade coseno < 0.5 para textos não relacionados |
| AC21 | Cache de embedding deve funcionar | Segunda chamada com mesmo input é instantânea |

## Rerank

| # | Critério | Como Verificar |
|---|----------|----------------|
| AC22 | Rerank deve retornar documentos ordenados | Documento mais relevante aparece primeiro |
| AC23 | top_n deve limitar resultados | top_n=2 retorna apenas 2 documentos |
| AC24 | Scores devem estar entre 0 e 1 | Todos scores no intervalo [0, 1] |

## API

| # | Critério | Como Verificar |
|---|----------|----------------|
| AC25 | Todos endpoints devem retornar JSON válido | Content-Type: application/json |
| AC26 | Erros devem seguir formato padrão | Erro contém code, message, request_id |
| AC27 | Health check deve retornar 200 | GET /health retorna status ok |
| AC28 | Request ID deve ser refletida na resposta | X-Request-ID enviado = mesmo ID na resposta |

## Performance & Stability

| # | Critério | Como Verificar |
|---|----------|----------------|
| AC29 | Sistema deve inicializar em < 2s | Tempo do start até primeiro health check 200 |
| AC30 | Cache deve reduzir tempo de resposta | Resposta com cache > 2x mais rápida que sem cache |
| AC31 | Múltiplas requisições concorrentes não devem quebrar | 10 requests paralelas → todas retornam 200 |
| AC32 | Graceful shutdown deve descarregar modelos | SIGTERM → modelos descarregados antes de exit |

## Testing

| # | Critério | Como Verificar |
|---|----------|----------------|
| AC33 | Todos os testes unitários devem passar | go test ./... não falha |
| AC34 | Cobertura mínima de 90% | go test -cover mostra >= 90% |
| AC35 | Lint não deve apontar erros | golangci-lint run não reporta erros |
| AC36 | Nenhuma dependência circular | go vet ./... não reporta erros |

## Documentation

| # | Critério | Como Verificar |
|---|----------|----------------|
| AC37 | Toda função exportada deve ter GoDoc | godoc não mostra warnings |
| AC38 | API deve ter documentação OpenAPI | Endpoint /openapi.json retorna spec |
| AC39 | README deve explicar instalação e uso | README contém exemplos funcionais |
| AC40 | ADRs devem refletir decisões atuais | ADRs em docs/adr/ estão atualizados |
