package main

import (
	"strings"

	"github.com/spf13/cobra"
)

// completeTags completes the segment after the last comma, so a partial
// `--tags frontend,bu<TAB>` offers `frontend,bug` rather than replacing what is
// already typed. NoSpace keeps the cursor where another comma can follow.
func completeTags(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	prefix, partial := "", toComplete
	if cut := strings.LastIndex(toComplete, ","); cut >= 0 {
		prefix, partial = toComplete[:cut+1], toComplete[cut+1:]
	}

	chosen := map[string]bool{}
	for _, tag := range strings.Split(prefix, ",") {
		chosen[strings.TrimSpace(tag)] = true
	}

	suggestions := make([]string, 0, len(knownTags()))
	for _, tag := range knownTags() {
		if chosen[tag] || !strings.HasPrefix(tag, partial) {
			continue
		}
		suggestions = append(suggestions, prefix+tag)
	}
	return suggestions, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
}

// registerTagCompletion wires completion onto a command's --tags flag. Errors
// are ignored deliberately: a missing flag is a programming error the build
// would not catch, but a shell that cannot complete is not worth failing a
// command over.
func registerTagCompletion(cmd *cobra.Command) {
	_ = cmd.RegisterFlagCompletionFunc("tags", completeTags)
}
