#!/usr/bin/env python3

import re
import unittest
from pathlib import Path


SOURCE = (
    Path(__file__).resolve().parents[1] / "proto" / "identity" / "v1" / "tokens.proto"
)


class IdentityWorkspaceCompatibilityTests(unittest.TestCase):
    def test_workspace_fields_match_platform_numbers(self) -> None:
        source = SOURCE.read_text(encoding="utf-8")
        issue = re.search(
            r"message IssueServiceTokenRequest \{(?P<body>.*?)\n\}", source, re.DOTALL
        )
        introspect = re.search(
            r"message IntrospectResponse \{(?P<body>.*?)\n\}", source, re.DOTALL
        )
        self.assertIsNotNone(issue)
        self.assertIsNotNone(introspect)
        self.assertRegex(issue.group("body"), r"\bstring workspace_id = 5;")
        self.assertRegex(introspect.group("body"), r"\bstring workspace_id = 27;")


if __name__ == "__main__":
    unittest.main()
