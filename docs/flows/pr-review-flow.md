# Flujo: PR Review

Flujo focalizado para revisión de PRs y diffs.

## Trigger

El usuario ejecuta `/sequoia review` con flags opcionales.

## Diagrama de flujo

```
/sequoia review [--diff=X] [--pr=N] [--strict]
  │
  ├─ 1. OBTENER DIFF
  │     ├─ --diff=HEAD~3..HEAD → git diff --stat + contenido
  │     ├─ --pr=42 → gh pr diff 42
  │     └─ default → git diff HEAD~1..HEAD
  │
  ├─ 2. CLASIFICAR ARCHIVOS
  │     └─ Mapear cada archivo cambiado a tipos de agentes
  │
  ├─ 3. AUTO-SELECT AGENTES
  │     ├─ Consultar tabla de mapeo archivo→agente
  │     └─ Mínimo: P3 Architecture (siempre corre)
  │
  ├─ 4. RECUPERAR HALLAZGOS PREVIOS
  │     ├─ Buscar en Engram hallazgos que afecten los archivos del diff
  │     └─ Marcar hallazgos previos que los cambios podrían resolver o afectar
  │
  ├─ 5. EJECUTAR AGENTES (solo sobre archivos del diff)
  │     ├─ Todos en paralelo (no hay dependencias en review)
  │     └─ Cada agente solo analiza los archivos cambiados
  │
  ├─ 6. GENERAR OUTPUT FOCALIZADO
  │     ├─ Hallazgos nuevos del diff
  │     ├─ Hallazgos previos afectados/resueltos
  │     └─ Veredicto final
  │
  └─ 7. PERSISTIR
        └─ Guardar hallazgos del review en Engram (separados de audit)
```

## Tabla de auto-selección: tipo de archivo → agentes

| Patrón de archivo | Agentes activados | Justificación |
|-------------------|-------------------|---------------|
| `**/auth/**`, `**/session/**`, `**/middleware/**` | P1 | Cambios en autenticación siempre son security-sensitive |
| `**/*.jsx`, `**/*.tsx`, `**/*.vue`, `**/*.svelte`, `**/*.html` | P2, P5 | Componentes UI: performance + experiencia |
| `**/api/**`, `**/routes/**`, `**/controllers/**`, `**/handlers/**` | P1, P3 | Endpoints: seguridad + arquitectura + API design |
| `**/models/**`, `**/schema/**`, `**/migrations/**`, `**/entities/**` | P3, P6 | Datos: arquitectura + operaciones |
| `**/*.test.*`, `**/*.spec.*`, `**/__tests__/**` | P4 | Tests: calidad |
| `Dockerfile*`, `**/.github/**`, `**/deploy/**`, `docker-compose.*` | P6 | Infra: operaciones |
| `package.json`, `go.mod`, `Cargo.toml`, `requirements.txt` | P4 | Deps: siempre revisar cambios en dependencias (parte de calidad) |
| `vite.config.*`, `webpack.config.*`, `tsconfig.*` | P2, P3 | Build config: performance + arquitectura |
| `*.css`, `*.scss`, `*.less` | P2, P5 | Estilos: performance + experiencia |
| `**/*.md` | — | Documentación: no audita (a menos que sea API docs) |
| **Cualquier otro** | P3 | Architecture siempre tiene algo que decir |

## Reglas de selección

1. **P3 Architecture siempre corre** — todo cambio tiene impacto arquitectural
2. **P4 Quality corre si cambió el manifiesto** — nuevas deps = nuevo riesgo
3. **P1 Security corre si hay auth, endpoints, o input handling** — seguridad no negocia
4. **P5 Experience solo si hay UI** — si el diff es solo backend, se salta
5. **Meta-agentes**: solo correlator simplificado, sin reporter completo

## Formato de output focalizado

```markdown
## Sequoia Review

**Rango**: [diff range o PR #N]
**Archivos**: {N} archivos cambiados
**Agentes**: {lista de agentes ejecutados}

---

### 🔴 Bloqueantes
{hallazgos críticos del diff}

### 🟠 Riesgos
{hallazgos de riesgo del diff}

### 🟡 Atención [--strict only]
{hallazgos medios, solo en modo estricto}

---

### 📋 Hallazgos previos afectados

| Hallazgo previo | Estado | Detalle |
|----------------|--------|---------|
| [P1-002] Token sin expiración | 🔸 Parcialmente resuelto | Se agregó expiración pero falta rotation |
| [P3-005] God module auth | ⏸️ No tocado | El diff no modifica auth/index.ts |

### ✅ Hallazgos resueltos por este diff
{si el cambio corrige un hallazgo previo}

---

**Veredicto**: {✅ PASS | ⚠️ WARN | 🔴 BLOCK}
**Revisión sugerida**: {comentario opcional sobre el cambio en conjunto}
```

## Lógica de veredicto

| Condición | Veredicto |
|-----------|-----------|
| Sin hallazgos 🔴 ni 🟠 | ✅ PASS |
| Solo hallazgos 🟠 y `--strict` no activo | ⚠️ WARN |
| Hallazgos 🔴 | 🔴 BLOCK |
| Hallazgos 🟠 y `--strict` activo | 🔴 BLOCK |
| Sin hallazgos pero hallazgos previos empeorados | ⚠️ WARN |

## Integración con hallazgos previos

1. **Recuperar** de Engram los hallazgos que citan archivos del diff
2. **Verificar** si las líneas cambiadas están cerca de hallazgos previos
3. **Marcar** cada hallazgo previo como:
   - `✅ Posiblemente resuelto` — si el diff toca las líneas citadas en la evidencia
   - `🔸 Parcialmente afectado` — si el diff toca el archivo pero no las líneas exactas
   - `🔻 Empeorado` — si el diff agrava el problema
   - `⏸️ No tocado` — si el diff no relaciona
4. Solo reportar los marcados como resuelto, afectado o empeorado (reducir ruido)

## Flags en áreas con hallazgos abiertos

Si el diff toca archivos con hallazgos 🔴 o 🟠 abiertos, agregar al output:

> ⚠️ **Este diff modifica áreas con hallazgos abiertos:**
> - `src/auth/handler.ts` tiene [P1-002] 🔴 Token sin expiración
> - `src/api/routes.ts` tiene [P3-001] 🟠 Endpoints sin paginación
>
> Verificar que los cambios no agraven estos hallazgos.
