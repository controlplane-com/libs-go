# Schema inventory & ownership map

Every non-test, non-`interfaces/` file in `nodelibs/schema/src/`, its role
(top-level **registered** schema vs shared **sub-schema**), which resource
kind(s)/partitions consume it, and its **exclusive** owner. Cross-partition
references are called out so a fan-out agent imports the owner's symbol instead of
re-authoring it.

Partitions: **A** workload · **B** IAM · **C** mk8s+cloud · **D** org+gvc ·
**E** data+net · **F** misc+query. **base.ts is DONE** (foundation,
`schema_base.go`).

Legend for "Role": `TOP` = registered per-kind schema (validated on create/patch);
`SUB` = shared sub-schema referenced by others; `LEAF` = helper/utility (no Joi
resource).

---

## Foundation (DONE — do not re-own)

| File | Role | Notes |
|------|------|-------|
| base.ts | SUB | Ported → `pkg/handlers/schema_base.go`. Symbols: Name, Kind, CloudProvider, Tags, description, Link/Links, LocalLink, ImageLink, ToKind, Regex, Base, checkImageAndTag. Everyone extends `BaseSchema()`. |
| internal.ts | LEAF | Internal-service alias registry (`keda`). Already ported into `links.go` (`internalAliases`, `validInternalLink`). Owner: **F** (note only; no new work unless registry grows). |
| links.ts | SUB | Joi `Links`/`Link` array schema. Superseded by `schema_base.go` `LinksRule()`/`LinkRule()` + the central `links.go` expansion. Owner: **F** (note only). |

---

## A / workload

| File | Role | Consumed by | Owner | Cross-refs (import, don't fork) |
|------|------|-------------|-------|------------------|
| workload.ts | TOP `workload` | gvc(D), command(E), mk8sAddons(C), mk8sEphemeral(C), valkey(E), volumeSet(E) | **A** | uses env, envoy, port, util, vm, volumeSpec, workloadOptions (all A/F below) |
| workloadOptions.ts | SUB | workload(A), serviceMapHelpers(A) | **A** | uses util(F) |
| port.ts | SUB (Port number rule) | workload(A), vm(A), envoyExcExtAuth(A), **domain(E)** | **A** | **domain (E) consumes Port — E imports A's `PortRule`** |
| env.ts | SUB (EnvVar/EnvValue) | workload(A), volumeSpec(A), **gvc(D)**, **command(E)** | **A** | double-listed in task (A and "F if shared"); resolved to **A** as the workload env model. D & E import A's `EnvVarRule`. If the launcher prefers a neutral home, move to F — flagged. |
| containerstatus.ts | SUB | deployment(A), cronjob(A) | **A** | — |
| cronjob.ts | SUB | deployment(A) | **A** | **NOT in the task's partition lists** — it is a deployment sub-schema; assigned to A. uses containerstatus(A). |
| deployment.ts | TOP `deployment` | — | **A** | uses containerstatus, cronjob, statefulSet (A) |
| statefulSet.ts | SUB | deployment(A) | **A** | **uses volumeSet (E)** — A imports E's `VolumeSet` spec symbol |
| vm.ts | SUB (workload VM) | workload(A) | **A** | uses port(A) |
| volumeSpec.ts | SUB | workload(A), envoy(A) | **A** | uses env(A) |
| envoy.ts | SUB | workload(A), **gvc(D)** | **A** | condition "only workload/gvc consume" holds → A. uses envoyAccessLog/Cluster/ExcExtAuth/Http, volumeSpec |
| envoyCommon.ts | SUB | envoy* (A) | **A** | — |
| envoyAccessLog.ts | SUB | envoy(A) | **A** | uses envoyCommon |
| envoyCluster.ts | SUB | envoy(A) | **A** | uses envoyCommon |
| envoyExcExtAuth.ts | SUB | envoy(A) | **A** | uses envoyCommon, port(A) |
| envoyHttp.ts | SUB | envoy(A) | **A** | uses envoyCommon |
| serviceMap.ts | SUB | serviceMapHelpers(A) | **A** | — |
| serviceMapHelpers.ts | SUB | (handlers) | **A** | uses workloadOptions(A), links(F) |

Note: envoy files contain many **commented-out** `Joi.link('#…')` recursive
schemas — they are NOT active; do not implement recursion for them.

## B / IAM

| File | Role | Consumed by | Owner | Cross-refs |
|------|------|-------------|-------|-----------|
| identity.ts | TOP `identity` | — | **B** | link fields already in `linkRegistry` |
| secret.ts | TOP `secret` | — | **B** | callsite uses allowUnknown (secret.ts:883/907) → `Prefs{AllowUnknown:true}` |
| policy.ts | TOP `policy` | — | **B** | uses query(F) |
| group.ts | TOP `group` | — | **B** | uses query(F) |
| serviceaccount.ts | TOP `serviceaccount` | — | **B** | — |
| user.ts | TOP `user` | — | **B** | — |
| policyMembership.ts | TOP `policymembership` | — | **B** | — |
| permissions.ts | SUB (permissions catalog) | — | **B** | — |
| resourcePolicy.ts | TOP `resourcepolicy` | — | **B** | — |
| accessreport.ts | TOP `accessreport` | — | **B** | — |

## C / mk8s + cloud

| File | Role | Consumed by | Owner | Cross-refs |
|------|------|-------------|-------|-----------|
| mk8s.ts | TOP `mk8s` | — | **C** | uses mk8sAddons + all providers |
| mk8sCommon.ts | SUB | all mk8s providers | **C** | — |
| mk8sAddons.ts | SUB | mk8s(C) | **C** | **uses workload(A)** — import A's symbol |
| mk8sAws.ts | SUB | mk8s(C) | **C** | uses mk8sCommon |
| mk8sAzure.ts | SUB | mk8s(C) | **C** | uses mk8sCommon |
| mk8sDigitalOcean.ts | SUB | mk8s(C) | **C** | uses mk8sCommon |
| mk8sEphemeral.ts | SUB | mk8s(C) | **C** | uses mk8sCommon, **workload(A)** |
| mk8sGcp.ts | SUB | mk8s(C) | **C** | uses mk8sCommon |
| mk8sGeneric.ts | SUB | mk8s(C) | **C** | uses mk8sCommon |
| mk8sHetzner.ts | SUB | mk8s(C) | **C** | uses mk8sCommon |
| mk8sLambdalabs.ts | SUB | mk8s(C) | **C** | uses mk8sCommon |
| mk8sLinode.ts | SUB | mk8s(C) | **C** | uses mk8sCommon |
| mk8sOblivus.ts | SUB | mk8s(C) | **C** | uses mk8sCommon |
| mk8sPaperspace.ts | SUB | mk8s(C) | **C** | uses mk8sCommon |
| mk8sTriton.ts | SUB | mk8s(C) | **C** | uses mk8sCommon |
| cloudaccount.ts | TOP `cloudaccount` | — | **C** | data = `alternatives().conditional('provider', …)` — use `Alternatives().Conditional` |
| agent.ts | TOP `agent` | — | **C** | — |

## D / org + gvc

| File | Role | Consumed by | Owner | Cross-refs |
|------|------|-------------|-------|-----------|
| org.ts | TOP `org` | — | **D** | uses orgLogging(D), tracing(D) |
| gvc.ts | TOP `gvc` | — | **D** | **uses env(A), envoy(A), workload(A)**, query(F), tracing(D) |
| orgLogging.ts | SUB | org(D) | **D** | logCfg `.xor(...)` over 11 providers |
| tracing.ts | SUB | org(D), gvc(D) | **D** | — |
| quota.ts | TOP `quota` | — | **D** | — |
| task.ts | TOP `task` | — | **D** | — |
| tenant.ts | TOP `tenant` | — | **D** | — |
| location.ts | TOP `location` | — | **D** | — |
| locationDns.ts | SUB | (handlers) | **D** | not imported by other schema files |
| auditctx.ts | TOP `auditctx` | — | **D** | uses query(F) |
| event.ts | TOP `event` | — | **D** | — |

## E / data + net

| File | Role | Consumed by | Owner | Cross-refs |
|------|------|-------------|-------|-----------|
| domain.ts | TOP `domain` | — | **E** | uses links(F), **port(A)**; link fields in `linkRegistry` |
| ipSet.ts | TOP `ipset` | — | **E** | link fields in `linkRegistry` |
| command.ts | TOP `command` | — | **E** | **uses env(A), workload(A)**, query(F), **volumeSet(E)** |
| memcache.ts | TOP `memcachecluster` | — | **E** | — |
| volumeSet.ts | TOP `volumeset` | **statefulSet(A)**, command(E) | **E** | A imports E's spec symbol |
| dbcluster.ts | TOP `dbcluster` | — | **E** | — |
| valkey.ts | TOP (valkey cluster) | — | **E** | **uses workload(A)** |
| spicedb.ts | TOP `spicedbcluster` | — | **E** | — |

## F / misc + query

| File | Role | Consumed by | Owner | Cross-refs |
|------|------|-------------|-------|-----------|
| image.ts | TOP `image` | — | **F** | image-string validation lives in `schema_base.go` `checkImageAndTag` |
| query.ts | SUB (query DSL) | auditctx(D), group(B), gvc(D), policy(B) | **F** | multiple B/D consumers import F's `QueryRule` |
| config.ts | TOP (config) | — | **F** | — |
| discovery.ts | LEAF | — | **F** | — |
| util.ts | LEAF (makeUnique/deleteFields helpers) | workload(A), workloadOptions(A) | **F** | `makeUnique` is `.Unique()` in the engine; `deleteFields*` are custom coercions — port as `CustomFn` |
| env.ts | — | — | **A** | (owned by A — see A table; listed here only because task wrote "env(if shared)") |
| internal.ts | LEAF | base(F) | **F** | already in `links.go` (note only) |
| links.ts | SUB | base(F) | **F** | superseded by `schema_base.go` + `links.go` (note only) |

---

## Flagged for launcher review

1. **cronjob.ts** is not named in any partition list; it is a `deployment`
   sub-schema → assigned to **A**. Confirm.
2. **env.ts** is double-listed (A explicitly, and "F if shared"). It IS shared
   (A/D/E). Resolved to **A** (workload env model); D & E import A's symbol. Flip
   to F if you prefer a neutral home — one-line change of owner.
3. **port.ts** owned by A but consumed by **domain (E)**; **volumeSet.ts** owned
   by E but consumed by **statefulSet (A)**; **workload/env/envoy** (A) consumed
   by gvc (D), command (E), mk8s (C), valkey (E). These are the only
   cross-partition symbol imports — all one-directional, no cycles.
4. **valkey.ts** / **spicedb.ts** map to kinds (`authzedcluster`/valkey) whose Go
   handlers are only partially mounted; the schema still ports fully (register on
   the kind even if the route lands later — `linkRegistry` already has stub
   entries).
5. **config.ts / discovery.ts** are thin; verify they are actually
   create/patch-validated resources before registering (some are read-only).
