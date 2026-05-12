---
description: "Inicializa Sequoia en el proyecto. Detecta stack, construye el mapa del proyecto, identifica agentes aplicables y persiste contexto en Engram. Primer paso obligatorio antes de cualquier auditoría."
allowed-tools: Read, Glob, Grep, Bash
---

# /sequoia init

Inicializa Sequoia en el proyecto actual. Construye el mapa completo que informa a TODOS los agentes posteriores.

## Qué hace

1. Ejecuta el agente de contexto (`sequoia-context`) como pre-flight
2. Construye el **Project Map** — la única fuente de verdad sobre el proyecto
3. Determina qué agentes aplican y cuáles no (con motivo explícito)
4. Persiste el mapa en Engram para que futuras sesiones lo recuperen

## Flujo de trabajo paso a paso

### Paso 1 — Escanear estructura del proyecto

```
Buscar:
├── Manifiestos de dependencias: package.json, go.mod, Cargo.toml, pom.xml,
│   requirements.txt, pyproject.toml, Gemfile, *.csproj, composer.json
├── Configuraciones: tsconfig.*, vite.config.*, webpack.config.*, next.config.*,
│   docker-compose.*, Dockerfile, Makefile, build.gradle, .env*, *.yaml, *.toml
├── Frameworks: next/, pages/, src/app/, angular.json, nuxt.config.*
├── CI/CD: .github/workflows/, .gitlab-ci.yml, Jenkinsfile, .circleci/
└── Documentación: README*, CHANGELOG*, docs/, CONTRIBUTING*
```

### Paso 2 — Analizar tech stack

Identificar con evidencia:
- **Lenguaje principal** (y secundarios si hay)
- **Framework** (React, Next.js, Express, Django, Spring, Gin, etc.)
- **Runtime** (Node, Deno, Bun, Go, Python, Java, .NET)
- **Bundler/Build** (Vite, Webpack, esbuild, Rollup, Turbopack, o ninguno)
- **Test runner** (Jest, Vitest, pytest, Go testing, xUnit, o ausente)
- **Package manager** (npm, pnpm, yarn, pip, cargo, go modules)

### Paso 3 — Determinar paradigma del proyecto

Clasificar en UNO principal (puede tener secundario):
- SPA | SSR | SSG | API REST | API GraphQL | CLI | Library | Monolito
- Microservicios | Fullstack | Mobile | Desktop | Plugin

Justificar la clasificación con evidencia del repo.

### Paso 4 — Estimar tamaño del proyecto

```markdown
| Métrica | Valor | Cómo se midió |
|---------|-------|---------------|
| LOC estimado | ~N | conteo de archivos × promedio estimado |
| Módulos principales | N | directorios con responsabilidad propia |
| Dependencias directas | N | del manifiesto de deps |
| Dependencias totales | ~N | del lockfile si existe |
```

### Paso 5 — Verificar infraestructura existente

| Aspecto | Presente | Parcial | Ausente |
|---------|----------|---------|---------|
| Tests unitarios | | | |
| Tests de integración | | | |
| CI/CD pipeline | | | |
| Linting/Formatting | | | |
| Type checking | | | |
| Documentación técnica | | | |
| `.env.example` / env contract | | | |
| Docker / containerización | | | |

### Paso 6 — Evaluar madurez del proyecto

Criterios para clasificar:

- **Prototipo**: sin tests, sin CI, README mínimo, deps mínimas, estructura flat
- **Desarrollo activo**: algunos tests, CI básico, estructura en módulos, README con instrucciones
- **Producción**: tests > 50%, CI con gates, documentación operativa, monitoring, releases estables

### Paso 7 — Determinar agentes aplicables

Para cada agente P1-P6, decidir si aplica con motivo:

```markdown
| Agente | Aplica | Motivo |
|--------|--------|--------|
| P1 Security | ✅ | Todo proyecto necesita revisión de seguridad |
| P2 Performance | ✅ | [justificar según tipo] |
| P3 Architecture | ✅ | Siempre (incluye API design) |
| P4 Quality | ✅ | Siempre (incluye dependencias) |
| P5 Experience | ❌ | API pura sin interfaz de usuario |
| P6 Operations | ✅ | Siempre, con ajuste por madurez (incluye data) |
```

### Paso 8 — Persistir en Engram

Guardar como observación con:
- **title**: "Sequoia Project Map — {nombre-proyecto}"
- **topic_key**: `sequoia/{nombre-proyecto}/project-map`
- **type**: `architecture`
- **content**: el Project Map completo en formato markdown

Si ya existe un Project Map previo, comparar y notificar cambios significativos.

## Formato de salida: Project Map

```markdown
## Sequoia Project Map — {nombre}

**Fecha**: {fecha}
**Madurez**: {prototipo | desarrollo | producción}
**Paradigma**: {tipo principal}

### Stack detectado
- Lenguaje: {detectado}
- Framework: {detectado o "ninguno"}
- Runtime: {detectado}
- Bundler: {detectado o "ninguno"}
- Test runner: {detectado o "AUSENTE"}
- Package manager: {detectado}

### Tamaño
- LOC: ~{N}
- Módulos: {lista}
- Dependencias: {N} directas, ~{N} totales

### Infraestructura
- Tests: {presente | parcial | ausente}
- CI/CD: {presente | parcial | ausente}
- Docs: {presente | parcial | ausente}
- Lint: {presente | ausente}

### Agentes aplicables
{tabla del paso 7}

### Notas de contexto
- {cualquier detalle relevante que afecte el análisis}
```

## Manejo de detección ambigua

Cuando el stack no es claro:
1. Listar las opciones posibles con confianza estimada
2. Buscar evidencia adicional (imports en archivos fuente, scripts en manifiestos)
3. Si persiste la ambigüedad, declararla explícitamente: `[AMBIGUO: podría ser X o Y]`
4. NO adivinar. La ambigüedad declarada es mejor que la suposición silenciosa.

Si no se detecta ningún manifiesto de dependencias:
- Declarar: "No se detectó manifiesto de dependencias estándar"
- Listar qué archivos se buscaron y no se encontraron
- Evaluar si es un proyecto sin deps (scripts sueltos, código bare) o si falta información

## Chunking (proyectos grandes)

`/sequoia init` estima el tamaño del proyecto durante el Paso 4. Si el proyecto supera
~200 archivos fuente, el Project Map activa el modo **chunking**: el campo `chunks:` se
agrega al mapa con una lista de scopes, cada uno definido por patrones glob de archivos
(ej. `src/auth/**`, `src/api/**`).

Cuando un comando de auditoría (`audit`, `review`, `diff`, `fix`) detecta un Project Map
con `chunked: true`, procesa los chunks secuencialmente — ejecutando todos los agentes
aplicables sobre cada chunk antes de avanzar al siguiente. Esto mantiene el uso de tokens
dentro del presupuesto sin sacrificar cobertura.

Los chunks se definen por dominio funcional (no por tipo de archivo). El resultado es una
auditoría completa aunque el proyecto sea grande, ejecutada en pasos que el modelo puede
procesar sin pérdida de contexto.

## Precondición

No requiere precondiciones. Este es SIEMPRE el primer comando.

## Post-condición

El Project Map queda persistido en Engram. Los comandos `audit`, `review`, `diff` y `fix` lo consumen automáticamente.
