#!/bin/bash
# Stop hook: mechanical half of ~/.claude/rules/concision.md.
#
# The rule failed three times as prompt-only guidance (a memory amended twice, then a rules file), so
# this makes the harness check it. Catches stock sycophancy openers by literal match; the sibling
# `type: prompt` Stop handler judges the structural rules (recaps, narration, length) a grep can't see.
#
# Ceiling worth knowing: Stop fires AFTER the message is delivered, so this cannot make the first reply
# terse — it forces an immediate correction and applies pressure on later turns. There is no
# pre-delivery hook event for assistant text.
#
# Fails open everywhere: any missing input, unwritable marker, or parse problem exits 0 rather than
# risking a blocked session.

INPUT=$(cat /dev/stdin)

# Provided directly on Stop input — no transcript parsing, so mid-turn text and previous turns can't
# be mistaken for this turn's reply.
LAST=$(printf '%s' "$INPUT" | jq -r '.last_assistant_message // empty' 2>/dev/null)
SESSION=$(printf '%s' "$INPUT" | jq -r '.session_id // "nosession"' 2>/dev/null)
[ -z "$LAST" ] && exit 0

# Strip fenced blocks, inline code, and double-quoted spans first: quoting the user, a reviewer, or a
# test fixture that contains a banned phrase must not trip the check. Also strips talking ABOUT the rule.
SCRUBBED=$(printf '%s' "$LAST" | sed -E -e '/^[[:space:]]*```/,/^[[:space:]]*```$/d' -e 's/`[^`]*`//g' -e 's/"[^"]*"//g')

# Only phrases with no legitimate use in a terse technical reply. Bare "right"/"good" are deliberately
# absent — they appear constantly in normal prose ("the right operand", "good defaults"). Trailing
# guards stop "to be fairly", "it's worthwhile", "fair_share_enabled" matching.
PATTERN="you( a|'?)re (absolutely |completely |entirely )?right|right to (push back|call (me )?out|be skeptical)"
PATTERN="$PATTERN|good (point|catch|question|call)|great (question|point|catch)|nice work"
PATTERN="$PATTERN|fair (enough|question|point)|to be fair([^a-z_]|$)|exactly right|your instinct( was)? right"
PATTERN="$PATTERN|that'?s the right call|i appreciate|thanks for (the|your|pointing)"
PATTERN="$PATTERN|worth noting|it'?s worth([^a-zA-Z]|$)|the honest answer|the one that mattered"
PATTERN="$PATTERN|absolutely[.,!]|of course([^a-z]|$)|notably[,]|importantly[,]|the significant one"
PATTERN="$PATTERN|i apologi[sz]e|sorry (about|for)|happy to help|let me know if"

HITS=$(printf '%s' "$SCRUBBED" | grep -oiE "$PATTERN" | sort -u | tr '\n' ';' | sed 's/;$//')
[ -z "$HITS" ] && exit 0

# One block per distinct reply, keyed on a hash of the text so a same-length rewrite is still checked
# and an identical re-Stop is not. Unwritable marker -> fail open (never risk a repeat-block loop).
DIGEST=$(printf '%s' "$LAST" | cksum | tr -d ' ')
MARKER="${TMPDIR:-/tmp}/claude-concision-$SESSION-$DIGEST"
[ -f "$MARKER" ] && exit 0
: > "$MARKER" 2>/dev/null || exit 0
find "${TMPDIR:-/tmp}" -maxdepth 1 -name 'claude-concision-*' -mtime +1 -delete 2>/dev/null

jq -n --arg hits "$HITS" '{
  decision: "block",
  reason: ("concision.md: resend without these phrases — " + $hits +
    ". Do not validate or flatter; state facts and stop.")
}'
