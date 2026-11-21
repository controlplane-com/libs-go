package envoyCluster

// EnumValues makes the generated string enums reflect to Joi.string().valid(...).
// The generator emits the constants but not the method, so without these the types
// validate as free-form strings.
func (ClusterType) EnumValues() []string {
	return []string{string(ClusterTypeStatic), string(ClusterTypeStrictDns), string(ClusterTypeLogicalDns), string(ClusterTypeEds), string(ClusterTypeOriginalDst)}
}
