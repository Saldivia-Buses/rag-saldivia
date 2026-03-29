# Convenciones de commits

## Formato

```
tipo(scope): descripción corta en inglés — planN fN
```

## Tipos

| Tipo | Cuándo |
|------|--------|
| `feat` | Nueva funcionalidad |
| `fix` | Corrección de bug |
| `refactor` | Cambio sin cambio de comportamiento |
| `style` | CSS/UI sin cambio de lógica |
| `test` | Agregar o modificar tests |
| `docs` | Documentación |
| `chore` | Mantenimiento (deps, config, CI) |
| `ci` | Cambios de CI/CD |

## Scopes (válidos en commitlint)

`setup`, `web`, `cli`, `db`, `config`, `logger`, `shared`, `auth`, `rag`,
`chat`, `admin`, `collections`, `ingestion`, `blackbox`, `deps`, `ci`,
`docs`, `changelog`, `plans`

## Ejemplos

```
feat(chat): integrate vercel ai sdk for streaming — plan14 f1
style(web): apply claude azure tokens to globals.css — plan15 f2
refactor(web): archive aspirational components — plan13 f3
docs(plans): update master plan with roadmap — plan13 f2
test(chat): add component tests for SourcesPanel — plan14 f3
chore(docs): rewrite all agents for typescript stack — plan13 f1
```

## Reglas

- Descripción siempre en inglés
- Siempre referenciar plan y fase al final
- Primera letra minúscula después del tipo
- Sin punto final
- Máximo 72 caracteres la primera línea
- Body opcional para contexto adicional
- Subject must be all lower-case (commitlint enforced)
