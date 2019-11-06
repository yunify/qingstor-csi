package neonsan

import "github.com/container-storage-interface/spec/lib/go/csi"

var DefaultVolumeAccessModeType = []csi.VolumeCapability_AccessMode_Mode{
	csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
}

var DefaultControllerServiceCapability = []csi.ControllerServiceCapability_RPC_Type{
	csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	//csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
	//csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
	//csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
	//csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
}

var DefaultNodeServiceCapability = []csi.NodeServiceCapability_RPC_Type{
	csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
	csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
	csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
}

var DefaultPluginCapability = []*csi.PluginCapability{
	{
		Type: &csi.PluginCapability_Service_{
			Service: &csi.PluginCapability_Service{
				Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
			},
		},
	},
	{
		Type: &csi.PluginCapability_VolumeExpansion_{
			VolumeExpansion: &csi.PluginCapability_VolumeExpansion{
				Type: csi.PluginCapability_VolumeExpansion_OFFLINE,
			},
		},
	},
	//{
	//	Type: &csi.PluginCapability_Service_{
	//		Service: &csi.PluginCapability_Service{
	//			Type: csi.PluginCapability_Service_VOLUME_ACCESSIBILITY_CONSTRAINTS,
	//		},
	//	},
	//},
}
