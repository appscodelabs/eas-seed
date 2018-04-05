/*
Copyright 2018 The Pharmer Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package internalversion

import (
	etcddiscovery "github.com/etcd-manager/etcd-discovery/apis/etcddiscovery"
	rest "k8s.io/client-go/rest"
)

// PingsGetter has a method to return a PingInterface.
// A group's client should implement this interface.
type PingsGetter interface {
	Pings() PingInterface
}

// PingInterface has methods to work with Ping resources.
type PingInterface interface {
	Create(*etcddiscovery.Ping) (*etcddiscovery.Ping, error)
	PingExpansion
}

// pings implements PingInterface
type pings struct {
	client rest.Interface
}

// newPings returns a Pings
func newPings(c *EtcddiscoveryClient) *pings {
	return &pings{
		client: c.RESTClient(),
	}
}

// Create takes the representation of a ping and creates it.  Returns the server's representation of the ping, and an error, if there is any.
func (c *pings) Create(ping *etcddiscovery.Ping) (result *etcddiscovery.Ping, err error) {
	result = &etcddiscovery.Ping{}
	err = c.client.Post().
		Resource("pings").
		Body(ping).
		Do().
		Into(result)
	return
}
