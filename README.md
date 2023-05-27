# Fly Machine Autoscaler

This autoscaler is stateless, it will pull the state from the API when it thinks there is a scaling condition. This means that no storage has to be attached, and you can move it around at will.

## Writing Policies

The `region` metric/label must always be returned in the series.

Errors will be thrown if queries come back without a series metric, and no scaling will occur.

You can limit the regions involved (and add new ones) with the `regions` parameter of the policy config.

If a known region is not found within a query and a minimum number of nodes are specified, then that region will immediately have its node count increased.

You can use this as a trick to add new regions quickly by simply creating a single machine, then settings regions and minimum in the scaling config and let the autoscaler bring up all regions to their minimum node count.

If a region is removed from the list, then it is ignored. It will no longer consider or modify a region that is in the autoscaling list, so that region will become orphaned from the autoscaler until it is added back.

## Dealing with volumes

As of now machines with volumes are not supported.

In the future, we will attempt to clone the machine with the same volume settings as found on an existing machine.# FlyMachineAutoscaler
