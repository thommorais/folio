// One shared symbol: a per-file `declare const brand` makes each file's
// Branded<string, 'PlanId'> a different type, so ids cannot cross modules.
declare const brand: unique symbol

export type Branded<T, B extends string> = T & { readonly [brand]: B }
