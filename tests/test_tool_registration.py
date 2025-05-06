import unittest
from unittest.mock import patch, MagicMock

from mcp_kubernetes.main import add_kubectl_tools, add_helm_tools
from mcp_kubernetes.security import SecurityConfig
from mcp_kubernetes.config import Config
from mcp_kubernetes.tool_registry import (
    KUBECTL_READ_ONLY_TOOLS,
    KUBECTL_RW_TOOLS,
    KUBECTL_ADMIN_TOOLS,
    HELM_READ_ONLY_TOOLS,
    HELM_RW_TOOLS,
    HELM_ADMIN_TOOLS,
)


class TestToolRegistration(unittest.TestCase):
    """Test tool registration functions."""

    def setUp(self):
        """Set up test environment."""
        self.mock_mcp = MagicMock()
        # Create a patch for the FastMCP instance
        self.patcher = patch("mcp_kubernetes.main.mcp", self.mock_mcp)
        self.patcher.start()

        # Set up configs
        self.config = Config()
        self.config.security_config = SecurityConfig()

    def tearDown(self):
        """Clean up after tests."""
        self.patcher.stop()

    def test_add_kubectl_tools_readonly(self):
        """Test that only read-only kubectl tools are registered in readonly mode."""
        # Configure readonly mode
        self.config.security_config.readonly = True

        # Call the function
        add_kubectl_tools(self.config)

        # Check that tool() was called once for each read-only tool
        self.assertEqual(self.mock_mcp.tool.call_count, len(KUBECTL_READ_ONLY_TOOLS))

        # Verify that each read-only tool was registered
        registered_tools = {
            call_args[0][0] for call_args in self.mock_mcp.tool().call_args_list
        }
        for tool in KUBECTL_READ_ONLY_TOOLS:
            self.assertIn(tool, registered_tools)

        # Verify that no RW or admin tools were registered
        for tool in KUBECTL_RW_TOOLS + KUBECTL_ADMIN_TOOLS:
            self.assertNotIn(tool, registered_tools)

    def test_add_kubectl_tools_non_readonly(self):
        """Test that all kubectl tools are registered in non-readonly mode."""
        # Configure non-readonly mode
        self.config.security_config.readonly = False

        # Call the function
        add_kubectl_tools(self.config)

        # Check that tool() was called once for each tool (read-only, RW, and admin)
        expected_tool_count = (
            len(KUBECTL_READ_ONLY_TOOLS)
            + len(KUBECTL_RW_TOOLS)
            + len(KUBECTL_ADMIN_TOOLS)
        )
        self.assertEqual(self.mock_mcp.tool.call_count, expected_tool_count)

        # Verify that all tools were registered
        registered_tools = {
            call_args[0][0] for call_args in self.mock_mcp.tool().call_args_list
        }
        for tool in KUBECTL_READ_ONLY_TOOLS + KUBECTL_RW_TOOLS + KUBECTL_ADMIN_TOOLS:
            self.assertIn(tool, registered_tools)

    def test_add_helm_tools_readonly(self):
        """Test that only read-only helm tools are registered in readonly mode."""
        # Configure readonly mode
        self.config.security_config.readonly = True

        # Call the function
        add_helm_tools(self.config)

        # Check that tool() was called once for each read-only tool
        self.assertEqual(self.mock_mcp.tool.call_count, len(HELM_READ_ONLY_TOOLS))

        # Verify that each read-only tool was registered
        registered_tools = {
            call_args[0][0] for call_args in self.mock_mcp.tool().call_args_list
        }
        for tool in HELM_READ_ONLY_TOOLS:
            self.assertIn(tool, registered_tools)

        # Verify that no RW or admin tools were registered
        for tool in HELM_RW_TOOLS + HELM_ADMIN_TOOLS:
            self.assertNotIn(tool, registered_tools)

    def test_add_helm_tools_non_readonly(self):
        """Test that all helm tools are registered in non-readonly mode."""
        # Configure non-readonly mode
        self.config.security_config.readonly = False

        # Call the function
        add_helm_tools(self.config)

        # Check that tool() was called once for each tool (read-only, RW, and admin)
        expected_tool_count = (
            len(HELM_READ_ONLY_TOOLS) + len(HELM_RW_TOOLS) + len(HELM_ADMIN_TOOLS)
        )
        self.assertEqual(self.mock_mcp.tool.call_count, expected_tool_count)

        # Verify that all tools were registered
        registered_tools = {
            call_args[0][0] for call_args in self.mock_mcp.tool().call_args_list
        }
        for tool in HELM_READ_ONLY_TOOLS + HELM_RW_TOOLS + HELM_ADMIN_TOOLS:
            self.assertIn(tool, registered_tools)


if __name__ == "__main__":
    unittest.main()
