export type RoutingSchemaForm = {
  code: string
  status: string
  isDefault: boolean
  schemaJson: string
}

export type RoutingMapperForm = {
  mapperType: string
  rulesJson: string
}

export type RoutingCompatibilityForm = {
  versionId: string
  inputSchemaId: string
  mapperId: string
  compatibilityType: string
  isDefaultRoute: boolean
  defaultValuesJson: string
}

export type RoutingExperimentForm = {
  name: string
  experimentType: string
  status: string
}

export type RoutingScopeForm = {
  experimentId: string
  inputSchemaId: string
  trafficPercent: number
  fallbackPolicy: string
  fallbackVersionId: string
  trafficConditionsJson: string
}

export type RoutingVariantForm = {
  scopeId: string
  versionId: string
  trafficWeight: number
  isControlGroup: boolean
  isActive: boolean
}

export type RoutingDeleteTarget = {
  message: string
  successTitle: string
  fallbackError: string
  action: () => Promise<unknown>
}
