# -*- coding: utf-8 -*-
import argparse
from fastmcp import FastMCP
import logging
from .kubeclient import setup_client, get
from .command import  helm
from .security import security_config


from .tool_registry import (
    KUBECTL_READ_ONLY_TOOLS,
    KUBECTL_RW_TOOLS,
    KUBECTL_ADMIN_TOOLS,
)

logger = logging.getLogger(__name__)


# Initialize FastMCP server
mcp = FastMCP("mcp-kubernetes")


def add_kubectl_tools():
    # Register read-only functions
    for func in KUBECTL_READ_ONLY_TOOLS:
        logger.debug(f"Registering kubectl function: {func.__name__}")
        mcp.tool()(func)

    # Register rw and admin functions
    if not security_config.readonly:
        for func in KUBECTL_RW_TOOLS + KUBECTL_ADMIN_TOOLS:
            logger.debug(f"Registering kubectl function: {func.__name__}")
            mcp.tool()(func)


def server():
    """Run the MCP server."""
    parser = argparse.ArgumentParser(description="MCP Kubernetes Server")

    # command options
    parser.add_argument(
        "--disable-helm",
        action="store_true",
        help="Disable helm command execution",
    )

    # Transport options
    parser.add_argument(
        "--transport",
        type=str,
        choices=["stdio", "sse"],
        default="stdio",
        help="Transport mechanism to use (stdio or sse)",
    )
    parser.add_argument(
        "--port",
        type=int,
        default=8000,
        help="Port to use for the server (only used with sse transport)",
    )

    # Security options
    parser.add_argument(
        "--readonly",
        action="store_true",
        default=False,
        help="Enable read-only mode (prevents write operations)",
    )
    parser.add_argument(
        "--allow-namespaces",
        type=str,
        default="",
        help="Comma-separated list of namespaces to allow (empty means all allowed)",
    )

    args = parser.parse_args()
    mcp.settings.port = args.port

    # Set security configuration
    security_config.readonly = args.readonly

    if args.allow_namespaces:
        security_config.allowed_namespaces = args.allow_namespaces

    # Setup Kubernetes client
    setup_client()

    # Setup kubectl tools
    # TODO: need a further discussion on using k8s sdk or kubectl, comment out these codes as they are duplicated with kubectl
    add_kubectl_tools()

    # Setup tools
    mcp.add_tool(get, name="Get-k8s-Object", description="Fetch any Kubernetes object (or list) as JSON string. Pass name='' to list the collection and namespace='' to get the resource in all namespaces.")
    if not args.disable_helm:
        mcp.tool("Run-helm-command","Run helm command and get result, The command should start with helm")(helm)

    # Run the server
    mcp.run(transport=args.transport)


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    server()
