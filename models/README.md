# Application models

The package groups reflect Dia's domains: `core` owns identity/session/account/
challenge/scene/role/group values, `system` owns node/key values and `comms` owns
communication delivery state. These types belong to Dia, not Toolbox.

`core.Ref` deliberately uses the public event-store root identity. Application
services validate transitions and authorization; storage adapters do not cast
or silently mutate these types. Wallet models remain cohesive with their
application service under `services/wallet`, using generic EVM primitives from
Toolbox. JSON field names and committed representations are unchanged by grouping.
