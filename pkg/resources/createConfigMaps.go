package resources

type CreateConfigMapOption interface {
	apply(*ConfigMapOptions) error
}

type createConfigMapOptionAdapter func(*ConfigMapOptions) error

func (c createConfigMapOptionAdapter) apply(co *ConfigMapOptions) error {
	return c(co)
}

func CreateCMWithName(name string) CreateConfigMapOption {
	return createConfigMapOptionAdapter(func(co *ConfigMapOptions) error {
		co.Name = name
		return nil
	})
}

func CreateCMWithData(data map[string]string) CreateConfigMapOption {
	return createConfigMapOptionAdapter(func(co *ConfigMapOptions) error {
		co.Data = data
		return nil
	})
}

func CreateCMInNamespace(namespace string) CreateConfigMapOption {
	return createConfigMapOptionAdapter(func(co *ConfigMapOptions) error {
		co.Namespace = namespace
		return nil
	})
}
