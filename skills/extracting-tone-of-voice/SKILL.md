---
name: extracting-tone-of-voice
description: >
  Analyze writing samples to extract a structured tone-of-voice profile that can be
  reused as a system prompt, style guide, or brand voice document. Use when the user
  wants to capture someone's writing style, create a tone-of-voice guide, clone a
  writing voice, build a style profile, analyze prose patterns, or standardize team
  communication style. Also use when the user says "write like me" or "match this tone".
license: MIT
metadata:
  author: jverhoeks
  version: "1.0.0"
  team: data-ai
---

# Extracting Tone of Voice

Analyze one or more writing samples to produce a structured, reusable tone-of-voice
profile. The output is a YAML profile that can be used as a system prompt ingredient,
a team style guide, or input for content generation tools.

## When to Use

- User provides writing samples and wants their style captured
- User says "write like me", "match this tone", "clone my voice"
- User wants to create a brand voice or style guide from examples
- User wants to standardize team writing across documents
- User wants to compare writing styles between authors

## Inputs Required

- **Minimum:** 1 writing sample (500+ words recommended)
- **Ideal:** 3-5 samples across different contexts (blog, email, docs, presentations)
- **For comparison:** Samples from 2+ authors labeled by author

More diverse samples produce a more accurate and robust profile.

## Analysis Procedure

### Step 1: Collect and Classify Samples

For each sample, note:
- Source type (blog post, email, documentation, presentation, chat message)
- Audience (technical peers, executives, general public, customers)
- Purpose (inform, persuade, instruct, entertain)
- Length (word count)

This context matters — a person writes differently in Slack vs. a white paper.
The profile should capture the *range*, not flatten it.

### Step 2: Analyze Structural Patterns

Examine these dimensions across all samples:

**Sentence Architecture**
- Average sentence length (short: <12 words, medium: 12-20, long: 20+)
- Sentence length variance (uniform vs. rhythmic mixing of short and long)
- Sentence openers: does the author lead with subject, clause, question, imperative?
- Use of fragments and one-word sentences
- Paragraph length tendencies

**Document Structure**
- How are ideas organized? (linear narrative, problem-solution, list-driven, layered)
- Use of headings and subheadings (heavy, light, none)
- Transition style between sections (explicit connectors, implicit, abrupt)
- Opening patterns (anecdote, question, bold claim, context-setting)
- Closing patterns (call to action, summary, open question, callback)

### Step 3: Analyze Language Choices

**Vocabulary and Register**
- Formality level (1-5 scale: 1=casual/slang, 5=academic/formal)
- Jargon usage: embraces domain terms, or translates for the reader?
- Vocabulary breadth: simple/accessible vs. varied/sophisticated
- Preferred verbs: active vs. passive voice ratio
- Use of first person (I/we) vs. third person vs. imperative

**Rhetorical Devices**
- Analogies and metaphors (frequent, occasional, rare)
- Type of analogies: technical, everyday, literary, humorous
- Rhetorical questions (and whether they're answered)
- Lists and enumeration style
- Use of contrast and juxtaposition
- Repetition for emphasis

**Personality Markers**
- Humor: absent, dry/subtle, playful, sarcastic
- Confidence level: hedging ("might", "perhaps") vs. assertive ("is", "must")
- Empathy markers: acknowledges reader's perspective? anticipates objections?
- Enthusiasm: measured/understated vs. energetic/emphatic
- Use of exclamation marks, em dashes, parentheticals, ellipses
- Emoji usage (never, sparingly, frequently)

### Step 4: Identify Signature Habits

Look for patterns unique to this author — things that would make their writing
recognizable without a byline:

- Characteristic phrases or expressions they reuse
- Distinctive punctuation habits (em dash lover, semicolon user, parenthetical asides)
- Preferred sentence templates ("The thing about X is...", "Here's the deal:")
- How they handle uncertainty ("I think" vs. "it seems" vs. just stating it)
- How they address the reader (direct "you", inclusive "we", impersonal)
- Cultural or domain references they draw from

### Step 5: Generate the Profile

Produce the profile in this YAML format:

```yaml
# Tone of Voice Profile
# Generated from [N] samples, [total word count] words
# Source types: [list]

profile:
  name: "Author Name / Brand Name"
  generated: "YYYY-MM-DD"
  sample_count: N
  total_words: NNNN

voice:
  formality: 3          # 1 (casual) to 5 (formal)
  confidence: 4         # 1 (hedging) to 5 (assertive)
  warmth: 3             # 1 (detached) to 5 (warm/personal)
  humor: 2              # 0 (none) to 5 (frequent/central)
  energy: 3             # 1 (calm/measured) to 5 (high-energy)
  technicality: 4       # 1 (layperson) to 5 (deep technical)

structure:
  avg_sentence_length: "medium"      # short / medium / long
  sentence_variance: "high"          # low / medium / high
  paragraph_length: "medium"         # short / medium / long
  preferred_openers:
    - "context-setting statement"
    - "direct claim"
  preferred_closers:
    - "actionable takeaway"
  organization: "layered"            # linear / problem-solution / layered / list-driven
  heading_usage: "moderate"          # none / light / moderate / heavy

language:
  voice: "active"                    # active / passive / mixed
  person: "first-plural"             # first-singular / first-plural / second / third / mixed
  jargon_handling: "uses-then-explains"  # avoids / uses-freely / uses-then-explains
  vocabulary_level: "accessible"     # simple / accessible / sophisticated / academic
  contraction_usage: "frequent"      # never / occasional / frequent

rhetoric:
  analogies: "frequent"              # rare / occasional / frequent
  analogy_domains:                   # where they draw analogies from
    - "military strategy"
    - "engineering"
  rhetorical_questions: "occasional"
  emphasis_method: "short-sentences" # bold / italics / caps / short-sentences / repetition

personality:
  humor_type: "dry"                  # dry / playful / sarcastic / self-deprecating / none
  punctuation_habits:
    - "heavy em dash usage"
    - "parenthetical asides"
  signature_phrases:
    - "Here's the thing:"
    - "The real question is"
  reader_address: "direct-you"       # direct-you / inclusive-we / impersonal
  uncertainty_style: "acknowledges-then-commits"  # hedges / omits / acknowledges-then-commits
  emoji: "never"                     # never / sparingly / frequently

# Context-dependent adjustments
registers:
  blog:
    formality_shift: -1              # relative to base
    humor_shift: +1
    notes: "More conversational, uses more analogies"
  email:
    formality_shift: 0
    notes: "Direct, shorter paragraphs, action-oriented"
  documentation:
    formality_shift: +1
    humor_shift: -1
    notes: "More structured, imperative mood, fewer opinions"
```

### Step 6: Generate the System Prompt Fragment

In addition to the YAML profile, produce a natural-language prompt fragment that can be
injected into any LLM system prompt:

```markdown
## Writing Style

Write in [Author]'s voice. Key characteristics:

- [2-3 sentence summary of overall tone]
- Sentence style: [describe rhythm and structure]
- Vocabulary: [describe register and jargon approach]
- Personality: [describe humor, confidence, warmth]
- Signature habits: [list 2-3 distinctive patterns]
- Address the reader as: [you/we/impersonal]
- Avoid: [2-3 things this author never does]
```

### Step 7: Validate (Optional)

If the user wants validation:
1. Generate a short paragraph (100-150 words) using the extracted profile
2. Ask the user: "Does this sound like you / the target author?"
3. Adjust the profile based on feedback
4. Repeat until the user confirms

## Output Files

Offer to save the profile in these formats:

| File | Purpose |
|------|---------|
| `tone-profile.yaml` | Machine-readable profile for tooling |
| `style-guide.md` | Human-readable style guide for teams |
| `system-prompt-fragment.md` | Copy-paste snippet for LLM prompts |

## Comparing Multiple Authors

When the user provides samples from 2+ authors:

1. Generate a separate profile for each
2. Produce a comparison table highlighting key differences:

```markdown
| Dimension       | Author A    | Author B    |
|-----------------|-------------|-------------|
| Formality       | 2 (casual)  | 4 (formal)  |
| Humor           | 3 (playful) | 0 (none)    |
| Sentence length | short       | long        |
| Jargon          | uses freely | explains    |
```

3. Note which author is better suited for which context

## Edge Cases

- **Too few words:** If samples total < 300 words, warn that the profile will be
  low-confidence. Still produce it, but mark dimensions as "uncertain" where evidence
  is thin.
- **Mixed authorship:** If a single document appears to have multiple authors (e.g.
  a co-written blog), flag this and ask the user which voice to extract.
- **Translated content:** Note that translation masks the original author's voice.
  The profile will reflect the translator's style as much as the author's.
- **Code-heavy samples:** Filter out code blocks before analyzing prose style.
  Note the ratio of prose to code.

## Tips

- Blog posts and long-form emails are the richest sources — they reveal natural voice
- Formal docs (RFCs, specs) show range but not personality
- Chat messages show personality but not structure
- The best profiles come from 3+ samples across different contexts
- Revisit the profile every 6-12 months — writing style evolves
