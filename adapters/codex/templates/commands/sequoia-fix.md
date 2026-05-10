---
description: "Genera plan de tareas accionables desde hallazgos de auditoría. Output optimizado para que otro agente implementador pueda ejecutar sin ambigüedad. Incluye contexto mínimo, archivos, criterio de aceptación."
argument-hint: "<phase|all> [--task=<id>]"
allowed-tools: Read, Glob, Grep
---

# /sequoia fix

Genera tareas implementables desde los hallazgos de una auditoría. Cada tarea es autosuficiente: un agente implementador puede ejecutarla sin releer la auditoría completa.

## Precondición

Debe existir al menos una auditoría previa en Engram (ejecutada via `/sequoia audit` o `/sequoia review`).

## Qué hace

1. Recupera hallazgos de la auditoría más reciente
2. Filtra por fase (si se especifica) o toma todas
3. Convierte cada hallazgo en una tarea implementable
4. Ordena por dependencias y prioridad
5. Genera el documento de tareas

## Uso

```bash
# Tareas de una fase específica
/sequoia fix security

# Tareas de todas las fases
/sequoia fix all

# Una tarea específica por ID
/sequoia fix security --task=P1-003
```

## Formato de cada tarea

Cada tarea generada sigue esta estructura obligatoria:

```markdown
### [TASK-ID] · [Título accionable]

**Prioridad**: Bloqueante | Alto leverage | Backlog
**Fase origen**: [P1-P6 | M1-M2]
**Hallazgo(s) origen**: [ID(s) del hallazgo que genera esta tarea]

**Contexto mínimo**:
Una explicación de QUÉ está mal y POR QUÉ importa, en 3-5 líneas.
Suficiente para entender el problema sin leer la auditoría completa.

**Archivos involucrados**:
- `path/al/archivo.ext` — qué papel cumple en esta tarea
- `path/otro/archivo.ext` — qué hay que modificar

**Qué hacer**:
Paso a paso concreto. No "mejorar X". Sino:
1. Agregar función Y en archivo Z
2. Modificar la llamada en archivo W para usar la nueva función
3. Actualizar el test en archivo T

**Impacto esperado**:
Qué cambia al implementar esto. Métrica observable si es posible.

**Dependencias**:
- Requiere que [TASK-ID] esté completa primero
- Bloqueado por: [factor externo, si aplica]

**Riesgo de implementación**: Bajo | Medio | Alto
Motivo del riesgo.

**Criterio de aceptación**:
- [ ] Condición verificable 1
- [ ] Condición verificable 2
- [ ] Test que debe pasar (si aplica)

**Verificación**:
Cómo confirmar que la tarea está realmente hecha.
Comando o paso manual concreto.
```

## Principio: tarea autosuficiente

Una tarea bien generada cumple estas reglas:

1. **No requiere leer la auditoría completa** — todo el contexto está en la tarea
2. **No es ambigua** — un desarrollador (o agente) puede implementar sin preguntas
3. **Tiene criterio de aceptación verificable** — no "mejorar X", sino "el test Y pasa"
4. **Declara dependencias explícitas** — sabe qué tareas deben ir antes
5. **Declara riesgo honestamente** — no todo es "riesgo bajo"

## Generación por fase vs todas

### Por fase (`/sequoia fix security`)
- Toma solo hallazgos de la fase indicada
- Ordena por severidad dentro de la fase
- Genera dependencias solo dentro de la fase

### Todas las fases (`/sequoia fix all`)
- Toma hallazgos de todas las fases
- Usa la correlación del M1 correlator para agrupar causas raíz
- Ordena globalmente: bloqueantes primero, luego alto leverage
- Genera dependencias cross-fase cuando la causa raíz es compartida

## Optimización del orden de implementación

Las tareas se ordenan siguiendo estos criterios:

1. **Bloqueantes de producción** → primero (sin excepción)
2. **Causas raíz** → antes que sus síntomas (del correlator)
3. **Dependencias técnicas** → si la tarea B requiere que A esté hecha
4. **Alto leverage** → máximo impacto con mínimo cambio
5. **Riesgo de implementación** → las de bajo riesgo antes (quick wins)

## Regla de deduplicación

Si múltiples hallazgos apuntan a la misma causa raíz (detectado por el correlator), se genera UNA tarea que resuelve todos los hallazgos relacionados. Se listan los IDs de hallazgo como origen.

## Output

Se genera `sequoia-fix.md` con la lista ordenada de tareas:

```markdown
# Plan de Implementación — [Proyecto]

**Generado desde**: Auditoría del [fecha]
**Total tareas**: {N}
**Bloqueantes**: {N} | **Alto leverage**: {N} | **Backlog**: {N}

## Orden de implementación
{tareas ordenadas por prioridad y dependencias}

## Dependencias entre tareas
{diagrama o lista de qué bloquea qué}

## Estimación de riesgo global
{evaluación del conjunto de cambios}
```

## Ejemplo

```bash
# Generar tareas de seguridad
/sequoia fix security

# Generar todas las tareas
/sequoia fix all

# Implementar una tarea específica
/sequoia fix security --task=P1-003
```
