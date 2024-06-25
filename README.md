# k8s-nas-maintenance

The main issue, this project solves for me, is the maintenance of the NAS that provides the storage for my k8s workloads (via the NFS provisioner). In order to correctly handle situations, in my opinion, the safe approach is to actually stop all workloads and unmount those mount points.

This project handles stopping / downscaling any k8s workloads, which mounts PVCs from the specific StorageClasses (e.g. the ones provided by NAS). Afterwards, it can upscale again and resume any paused Cronjobs.
