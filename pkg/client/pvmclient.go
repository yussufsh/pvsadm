package client

import (
	"fmt"
	"github.com/IBM-Cloud/bluemix-go/api/resource/resourcev2/controllerv2"
	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/client/image"
	"github.com/ppc64le-cloud/pvsadm/pkg/client/instance"
	"github.com/ppc64le-cloud/pvsadm/pkg/client/volume"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"k8s.io/klog/v2"
	"time"
)

type PVMClient struct {
	InstanceID string
	Region     string
	Zone       string

	PISession      *ibmpisession.IBMPISession
	InstanceClient *instance.Client
	ImgClient      *image.Client
	VolumeClient   *volume.Client
}

func NewPVMClient(c *Client, instanceID, instanceName string) (*PVMClient, error) {
	pvmclient := &PVMClient{}
	if instanceID == "" {
		svcs, err := c.ResourceClient.ListInstances(controllerv2.ServiceInstanceQuery{
			Type: "service_instance",
		})
		if err != nil {
			return pvmclient, fmt.Errorf("failed to list the resource instances: %v", err)
		}
		found := false
		for _, svc := range svcs {
			klog.V(4).Infof("Service ID: %s, region_id: %s, Name: %s", svc.Guid, svc.RegionID, svc.Name)
			klog.V(4).Infof("crn: %v", svc.Crn)
			if svc.Name == instanceName {
				instanceID = svc.Guid
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("instance: %s not found", instanceName)
		}
	}

	pvmclient.InstanceID = instanceID
	svc, err := c.ResourceClient.GetInstance(instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get a service with ID: %s, err: %v", instanceID, err)
	}

	pvmclient.Zone = svc.RegionID

	//re, err := regexp.Compile("[^a-zA-Z]+")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//region := re.ReplaceAllString(pvmclient.Zone, "")
	pvmclient.Region, err = utils.GetRegion(pvmclient.Zone)
	if err != nil {
		return nil, err
	}
	pvmclient.PISession, err = ibmpisession.New(c.Config.IAMAccessToken, pvmclient.Region, pkg.Options.Debug, 60*time.Minute, c.User.Account, pvmclient.Zone)
	if err != nil {
		return nil, err
	}

	pvmclient.ImgClient = image.NewClient(pvmclient.PISession, instanceID)
	pvmclient.VolumeClient = volume.NewClient(pvmclient.PISession, instanceID)
	pvmclient.InstanceClient = instance.NewClient(pvmclient.PISession, instanceID)
	return pvmclient, nil
}