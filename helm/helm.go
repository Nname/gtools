package helm

import (
	"context"
	syserror "errors"
	"fmt"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
)

type CdEngine struct {
	ApiServerURL     string
	Token            string
	ChartDir         string
	ReleaseName      string
	ReleaseNameSpace string
}

func (e *CdEngine) RestConfig() *rest.Config {
	config := &rest.Config{
		Host:        e.ApiServerURL,
		BearerToken: e.Token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
	return config
}

func (e *CdEngine) DynamicClient() (*dynamic.DynamicClient, error) {
	client, err := dynamic.NewForConfig(e.RestConfig())
	if err != nil {
		return nil, err
	}
	return client, err
}

func (e *CdEngine) RenderData(NewValues map[string]interface{}) (map[string]string, error) {
	chart, err := loader.Load(e.ChartDir)
	if err != nil {
		return nil, err
	}
	cValues, err := chartutil.CoalesceValues(chart, NewValues)
	if err != nil {
		return nil, err
	}
	options := chartutil.ReleaseOptions{
		Name:      e.ReleaseName,
		Namespace: e.ReleaseNameSpace,
	}
	valuesToRender, err := chartutil.ToRenderValues(chart, cValues, options, nil)
	render, err := engine.Render(chart, valuesToRender)
	if err != nil {
		return nil, err
	}
	return render, nil
}

func (e *CdEngine) DryRun(data []byte) error {
	client, err := e.DynamicClient()
	if err != nil {
		return err
	}
	utilruntime.Must(scheme.AddToScheme(scheme.Scheme))
	codecs := serializer.NewCodecFactory(scheme.Scheme)
	decoder := codecs.UniversalDeserializer()
	obj, _, err := decoder.Decode(data, nil, nil)
	if err != nil {
		return err
	}
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}
	resource := &unstructured.Unstructured{Object: unstructuredObj}
	gvr, _ := meta.UnsafeGuessKindToResource(resource.GroupVersionKind())
	_, err = client.Resource(gvr).Namespace(e.ReleaseNameSpace).Get(context.TODO(), resource.GetName(), v1.GetOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (e *CdEngine) Check(values map[string]interface{}) error {
	data, err := e.RenderData(values)
	if err != nil {
		return err
	}
	for yamlName, chart := range data {
		if err := e.DryRun([]byte(chart)); err != nil {
			return syserror.New(fmt.Sprintf("err, resource:[%s]", yamlName))
		}
	}
	return nil
}

func (e *CdEngine) Install(data []byte) error {
	client, err := e.DynamicClient()
	if err != nil {
		return err
	}
	utilruntime.Must(scheme.AddToScheme(scheme.Scheme))
	codecs := serializer.NewCodecFactory(scheme.Scheme)
	decoder := codecs.UniversalDeserializer()
	obj, _, err := decoder.Decode(data, nil, nil)
	if err != nil {
		return err
	}
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}
	resource := &unstructured.Unstructured{Object: unstructuredObj}
	gvr, _ := meta.UnsafeGuessKindToResource(resource.GroupVersionKind())
	_, err = client.Resource(gvr).Namespace(e.ReleaseNameSpace).Get(context.TODO(), resource.GetName(), v1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err := client.Resource(gvr).Namespace(e.ReleaseNameSpace).Create(context.TODO(), resource, v1.CreateOptions{})
		return err
	} else if err != nil {
		return syserror.New("the yaml resource file format is incorrect")
	} else {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			_, updateErr := client.Resource(gvr).Namespace(e.ReleaseNameSpace).Update(context.TODO(), resource, v1.UpdateOptions{})
			return updateErr
		})
		return retryErr
	}
}

func (e *CdEngine) Installs(values map[string]interface{}) error {
	data, err := e.RenderData(values)
	if err != nil {
		return err
	}
	for yamlName, chart := range data {
		if err := e.Install([]byte(chart)); err != nil {
			// fmt.Println(err)
			return syserror.New(fmt.Sprintf("err, resource:[%s], msg:[%s]", yamlName, err))
		}
	}
	return nil
}
