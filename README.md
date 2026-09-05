# Post-it 📝

Aplicativo ultra-rápido de notas adesivas (post-it) para macOS (Apple Silicon / Intel) e futuramente Linux, construído com foco em alta performance, minimalismo e estética natural de papel e caneta feita à mão.

Inspirado na arquitetura local-first e sem distrações do [task-space](https://github.com/MrSheerluck/task-space) e no design hand-drawn doodle do [drawably](https://github.com/Danilaa1/drawably).

---

## ✨ Funcionalidades Principais

* **Sobreposição em Todos os Apps (Always-on-Top):** Post-its flutuantes (`NSFloatingWindowLevel`) visíveis em todos os monitores e espaços (`Spaces`) do macOS, sem bloquear cliques em áreas vazias.
* **Arrasto Fluido e Nativo:** Cada post-it pode ser arrastado pelo topo com aceleração nativa de hardware da WindowServer do macOS.
* **Gesto Mágico de Balançar o Mouse:**
  * Balance o mouse rapidamente de um lado para o outro para ocultar todos os post-its instantaneamente.
  * Balance o mouse novamente para trazê-los de volta exatamente onde estavam.
* **Atalhos Globais do Sistema (Keycaps):**
  * `Cmd + Shift + N`: Criar uma nova nota adesiva.
  * `Cmd + Shift + P`: Alternar visibilidade (Ocultar / Exibir todas as notas).
* **Menu Suspenso no macOS (Barra de Status / Tray):**
  * Ícone `📝` na barra de menu com opções rápidas: Criar Nota, Ocultar/Exibir, Abrir Menu de Papéis e Sair.
  * Roda silenciosamente em segundo plano (`LSUIElement`) sem ocupar espaço desnecessário no Dock.
* **Tipos de Papel:**
  * **Pólen:** Tom suave e aconchegante amarelo pastel.
  * **Sulfite:** Base branca clássica, personalizável com seletor de cores e saturação.
  * **Couché:** Acabamento liso e sedoso com leve gradiente.
  * **Kraft:** Textura rústica e orgânica bege/marrom.
* **Pautas da Folha:**
  * **Pontilhado:** Grade de micropontos estilo Bullet Journal.
  * **Quadriculado:** Malha quadriculada suave para gráficos e organização.
  * **Linhas:** Linhas de caderno com entrelinha ajustada à escrita.
  * **Liso:** Folha pura sem pautas.
* **Estilo Caneta & Tipografia:**
  * Fonte Google **Caveat** integrada localmente (funciona 100% offline).
  * Cores de caneta: **Azul**, **Preto**, **Vermelho** e **Branco** (caneta gel para papéis escuros).
  * Alinhamentos de escrita: **Cantos** (início à esquerda), **Centralizada** e **Livre / Justificada**.
* **Controles Hand-Drawn (Drawably):**
  * Botões e controles com contornos desenhados à mão e efeito dinâmico doodle.
* **Persistência Local-First:**
  * Todas as notas e preferências são salvas automaticamente em `~/.config/post-it/notes.json` e `settings.json` com escritas atômicas seguras.

---

## 🏗️ Arquitetura e Engenharia de Software

O projeto segue princípios rigorosos de **Clean Architecture**, **Clean Code** e separação de responsabilidades (SOLID):

```
post-it/
├── cmd/
│   └── post-it/
│       └── main.go           # Ponto de entrada, runtime OS thread lock
├── internal/
│   ├── config/               # Gerenciador de caminhos e diretórios do app
│   ├── model/                # Tipos de domínio puros (Note, Settings, PaperType...)
│   ├── store/                # Persistência atômica thread-safe em disco
│   ├── platform/
│   │   └── macos/            # Integração nativa AppKit/WebKit, Carbon Hotkeys e Shake Detector
│   └── ui/                   # Templates HTML, CSS de papel, Caveat woff2 e Drawably
├── scripts/
│   └── bundle_app.sh         # Empacotador do .app do macOS
├── Makefile                  # Automação de compilação, testes e empacotamento
└── go.mod
```

---

## 🚀 Como Compilar e Rodar

### Pré-requisitos
* macOS (Apple Silicon M1/M2/M3/M4 ou Intel)
* Go 1.22+ instalado
* Clang / Xcode Command Line Tools (`xcode-select --install`)

### Comandos do Makefile

```bash
# Executar a suíte de testes unitários
make test

# Compilar o binário otimizado
make build

# Gerar o pacote Post-it.app completo
make app

# Abrir o aplicativo diretamente
make run
```

---

## 🛡️ Segurança e Performance

* **Zero chamadas de rede externas:** Nenhum dado sai do seu computador. Fontes, scripts e folhas rodam localmente da memória via `embed.FS`.
* **Binário ultraleve:** Tamanho inferior a 6 MB com tempo de inicialização imperceptível (< 50ms).
* **Consumo mínimo de CPU e memória:** O detector de movimento do mouse consulta as coordenadas em nanosegundos sem disparar permissões invasivas de acessibilidade do macOS.
