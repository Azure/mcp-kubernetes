from .kubectl import *
from .helm import *

# kubectl tools
KUBECTL_READ_ONLY_TOOLS = [
    kubectl_get,
    kubectl_describe,
    kubectl_explain,
    kubectl_logs,
    kubectl_api_resources,
    kubectl_api_versions,
    kubectl_diff,
    kubectl_cluster_info,
    kubectl_top,
    kubectl_events,
    kubectl_auth,
]

KUBECTL_RW_TOOLS = [
    kubectl_create,
    kubectl_delete,
    kubectl_apply,
    kubectl_expose,
    kubectl_run,
    kubectl_set,
    kubectl_rollout,
    kubectl_scale,
    kubectl_autoscale,
    kubectl_label,
    kubectl_annotate,
    kubectl_patch,
    kubectl_replace,
    kubectl_cp,
    kubectl_exec,
]

KUBECTL_ADMIN_TOOLS = [
    kubectl_cordon,
    kubectl_uncordon,
    kubectl_drain,
    kubectl_taint,
    kubectl_certificate,
]

KUBECTL_ALL_TOOLS = KUBECTL_READ_ONLY_TOOLS + KUBECTL_RW_TOOLS + KUBECTL_ADMIN_TOOLS


# helm tools
HELM_READ_ONLY_TOOLS = [
    helm_list,
    helm_get,
    helm_status,
    helm_history,
    helm_search,
    helm_template,
    helm_show,
    helm_verify,
    helm_env,
]

HELM_RW_TOOLS = [
    helm_install,
    helm_upgrade,
    helm_rollback,
    helm_uninstall,
    helm_test,
]

HELM_ADMIN_TOOLS = [
    helm_repo,
    helm_push,
    helm_dependency,
    helm_package,
    helm_registry,
    helm_pull,
]

HELM_ALL_TOOLS = HELM_READ_ONLY_TOOLS + HELM_RW_TOOLS + HELM_ADMIN_TOOLS
