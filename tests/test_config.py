"""Tests for the config module."""

import pytest

from spektacular.config import AgentConfig, DebugConfig, SpektacularConfig


class TestAgentConfig:
    def test_defaults(self):
        config = AgentConfig()
        assert config.command == "claude"
        assert "--output-format" in config.args
        assert "stream-json" in config.args
        assert "Bash" in config.allowed_tools
        assert config.dangerously_skip_permissions is False

    def test_custom_command(self):
        config = AgentConfig(command="my-agent")
        assert config.command == "my-agent"


class TestDebugConfig:
    def test_defaults(self):
        config = DebugConfig()
        assert config.enabled is False
        assert config.log_dir == ".spektacular/logs"

    def test_enabled(self):
        config = DebugConfig(enabled=True)
        assert config.enabled is True

    def test_custom_log_dir(self):
        config = DebugConfig(log_dir="/tmp/my-logs")
        assert config.log_dir == "/tmp/my-logs"


class TestSpektacularConfig:
    def test_has_agent_field(self):
        config = SpektacularConfig()
        assert isinstance(config.agent, AgentConfig)
        assert config.agent.command == "claude"

    def test_has_debug_field(self):
        config = SpektacularConfig()
        assert isinstance(config.debug, DebugConfig)
        assert config.debug.enabled is False

    def test_yaml_round_trip(self, tmp_path):
        config = SpektacularConfig()
        config_path = tmp_path / "config.yaml"
        config.to_yaml_file(config_path)

        loaded = SpektacularConfig.from_yaml_file(config_path)
        assert loaded.agent.command == config.agent.command
        assert loaded.agent.allowed_tools == config.agent.allowed_tools
        assert loaded.agent.dangerously_skip_permissions == config.agent.dangerously_skip_permissions
        assert loaded.debug.enabled == config.debug.enabled
        assert loaded.debug.log_dir == config.debug.log_dir

    def test_yaml_round_trip_debug_enabled(self, tmp_path):
        config = SpektacularConfig(debug=DebugConfig(enabled=True))
        config_path = tmp_path / "config.yaml"
        config.to_yaml_file(config_path)

        loaded = SpektacularConfig.from_yaml_file(config_path)
        assert loaded.debug.enabled is True
