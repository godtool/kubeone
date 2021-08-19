import Layout from "@/business/app-layout/horizontal-layout";

const Storage = {
  path: "/storage",
  sort: 5,
  parent: true,
  component: Layout,
  name: "Storage",
  meta: {
    title: "business.storage.storage",
    icon: "iconfont iconstorage",
  },
  children: [
    {
      path: "/persistentvolumes",
      requirePermission: {
        apiGroup: "",
        resource: "persistentvolumes",
        verb: "list",
      },
      component: () => import("@/business/storage/pv"),
      name: "PersistentVolumes",
      meta: {
        title: "Persistent Volume",
      },
    },
    {
      path: "/persistentvolumes/create",
      requirePermission: {
        apiGroup: "",
        resource: "persistentvolumes",
        verb: "create",
      },
      component: () => import("@/business/storage/pv/create"),
      name: "PersistentVolumeCreate",
      hidden: true,
      meta: {
        title: "Persistent Volume Create",
        activeMenu: "/persistentvolumes",
      }
    },
    {
      path: "/persistentvolumes/edit/:name",
      requirePermission: {
        apiGroup: "",
        resource: "persistentvolumes",
        verb: "update",
      },
      component: () => import("@/business/storage/pv/edit"),
      name: "PersistentVolumeEdit",
      props: true,
      hidden: true,
      meta: {
        title: "Persistent Volume Edit",
        activeMenu: "/persistentvolumes"
      }
    },
    {
      path: "/persistentvolumes/detail/:name",
      requirePermission: {
        apiGroup: "",
        resource: "persistentvolumes",
        verb: "get",
      },
      component: () => import("@/business/storage/pv/detail"),
      name: "PersistentVolumeDetail",
      props: true,
      hidden: true,
      meta: {
        title: "Persistent Volume Detail",
        activeMenu: "/persistentvolumes"
      }
    },
    {
      path: "/storageclasses",
      requirePermission: {
        apiGroup: "storage.k8s.io",
        resource: "storageclasses",
        verb: "list",
      },
      component: () => import("@/business/storage/sc"),
      name: "StorageClasses",
      meta: {
        title: "Storage Classes",
        activeMenu: "/storageclasses",
      }
    },
    {
      path: "/storageclasses/create",
      requirePermission: {
        apiGroup: "storage.k8s.io",
        resource: "storageclasses",
        verb: "crate",
      },
      component: () => import("@/business/storage/sc/create"),
      name: "StorageClassCreate",
      hidden: true,
      meta: {
        title: "Storage Classes Create",
        activeMenu: "/storageclasses",
      }
    },
    {
      path: "/storageclasses/edit/:name",
      requirePermission: {
        apiGroup: "storage.k8s.io",
        resource: "storageclasses",
        verb: "create",
      },
      component: () => import("@/business/storage/sc/edit"),
      name: "StorageClassEdit",
      hidden: true,
      props: true,
      meta: {
        title: "Storage Classes Edit",
        activeMenu: "/storageclasses",
      }
    },
    {
      path: "/storageclasses/detail/:name",
      requirePermission: {
        apiGroup: "storage.k8s.io",
        resource: "storageclasses",
        verb: "get",
      },
      component: () => import("@/business/storage/sc/detail"),
      name: "StorageClassDetail",
      props: true,
      hidden: true,
      meta: {
        title: "Storage Classes Detail",
        activeMenu: "/storageclasses",
      }
    }, {
      path: "/persistentVolumeClaim",
      requirePermission: {
        apiGroup: "",
        resource: "persistentvolumeclaims",
        verb: "list",
      },
      component: () => import("@/business/storage/pvc/"),
      name: "PersistentVolumeClaim",
      props: true,
      meta: {
        title: "Persistent Volume Claim",
        activeMenu: "/persistentVolumeClaim",
      }
    }, {
      path: "/persistentVolumeClaim/create",
      requirePermission: {
        apiGroup: "",
        resource: "persistentvolumeclaims",
        verb: "create",
      },
      component: () => import("@/business/storage/pvc/create"),
      name: "PersistentVolumeClaimCreate",
      hidden: true,
      meta: {
        title: "PersistentVolumeClaim Create",
        activeMenu: "/persistentVolumeClaim",
      }
    }, {
      path: "/:namespace/persistentVolumeClaim/detail/:name",
      requirePermission: {
        apiGroup: "",
        resource: "persistentvolumeclaims",
        verb: "get",
      },
      component: () => import("@/business/storage/pvc/detail"),
      name: "PersistentVolumeClaimDetail",
      hidden: true,
      props: true,
      meta: {
        title: "PersistentVolumeClaim Detail",
        activeMenu: "/persistentVolumeClaim",
      }
    },
  ]
}

export default Storage
