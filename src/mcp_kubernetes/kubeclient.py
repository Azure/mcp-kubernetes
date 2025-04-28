# -*- coding: utf-8 -*-
import base64
import json
import os
from datetime import datetime
from kubernetes import client, config, dynamic
from fastmcp import FastMCP
from pydantic import Field
from typing import Annotated


# Initialize FastMCP server
k8sapi = FastMCP("k8s-apiserver")

def gen_kubeconfig():
    """Generate a kubeconfig for the current Pod."""
    token = (
        open("/run/secrets/kubernetes.io/serviceaccount/token", "r", encoding="utf-8")
        .read()
        .strip()
    )  # Strip newline characters
    cert = (
        open("/run/secrets/kubernetes.io/serviceaccount/ca.crt", "r", encoding="utf-8")
        .read()
        .strip()
    )  # Strip newline characters
    cert = base64.b64encode(cert.encode()).decode()
    host = os.environ.get("KUBERNETES_SERVICE_HOST")
    port = os.environ.get("KUBERNETES_SERVICE_PORT")

    return f"""apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: {cert}
    server: https://{host}:{port}
  name: kube
contexts:
- context:
    cluster: kube
    user: kube
  name: kube
current-context: kube
kind: Config
users:
- name: kube
  user:
    token: {token}
"""


def setup_kubeconfig():
    """Set up kubeconfig if running inside a Pod."""
    if os.getenv("KUBECONFIG") is not None and os.getenv("KUBECONFIG") != "":
        return

    if not os.getenv("KUBERNETES_SERVICE_HOST"):
        # Not running inside a Pod, so no need to set up kubeconfig
        return

    home = os.path.expanduser("~")  # Use expanduser to get user's home directory
    kubeconfig_path = os.path.join(home, ".kube")
    kubeconfig_file = os.path.join(kubeconfig_path, "config")

    # If kubeconfig already exists, no need to recreate it
    if os.path.exists(kubeconfig_file):
        return

    os.makedirs(kubeconfig_path, exist_ok=True)
    kubeconfig = gen_kubeconfig()
    with open(kubeconfig_file, "w", encoding="utf-8") as f:
        f.write(kubeconfig)


class DateTimeEncoder(json.JSONEncoder):
    def default(self, o):
        if isinstance(o, datetime):
            return o.isoformat()
        return super().default(o)


def setup_client():
    """Get a Kubernetes client."""

    setup_kubeconfig()
    try:
        config.load_kube_config()
    except Exception:  # pylint: disable=broad-exception-caught
        config.load_incluster_config()
    return client

async def get(
    kind: Annotated[str, Field(description="The kubernetes resource kind")], 
    name: Annotated[str, Field(description="The kubernetes resource name, list all of resources if empty")] = "", 
    namespace: Annotated[str, Field(description="The kubernetes resource namespace, list all resource of all namespace if resource is namespace scoped and it is empty")] = "") -> bytes:
    """
    Fetch any Kubernetes object (or list) as JSON string. Pass name="" to list the collection and namespace="" to get the resource in all namespaces.

    :param kind: The resource type (e.g., pods, deployments).
    :param name: The name of the resource.
    :param namespace: The namespace of the resource.
    :return: The JSON representation of the resource.
    """
    api_client = client.ApiClient()
    dyn = dynamic.DynamicClient(api_client)
    fetched = list()
    resource = dyn.resources.get(kind=kind.capitalize())
    if resource.namespaced:
        if name:
            fetched = resource.get(name=name, namespace=namespace or "default")
        else:
            if namespace == "" or namespace is None:
                fetched = resource.get(all_namespaces=True)
            else:
                fetched = resource.get(namespace=namespace)
    else:
        fetched = resource.get(name=name) if name else resource.get()

    return json.dumps(fetched.to_dict(), indent=2, cls=DateTimeEncoder)


