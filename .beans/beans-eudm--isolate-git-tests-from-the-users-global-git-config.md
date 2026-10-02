---
# beans-eudm
title: Isolate git tests from the user's global git config
status: todo
type: bug
created_at: 2026-10-02T09:13:45Z
updated_at: 2026-10-02T09:13:45Z
---

TestMergeBase, TestMergeBase_FallbackToRemote and TestAllChangesVsUpstream_CommittedOnly fail when ~/.gitconfig sets push.pushOption: the local bare remote rejects push options ("the receiving end does not support push options"). The tests pass with GIT_CONFIG_GLOBAL=/dev/null.

- [ ] Set GIT_CONFIG_GLOBAL (and GIT_CONFIG_NOSYSTEM) for git commands run by the test helpers
