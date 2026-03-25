# Documentation Noob Testing

You are a brand new user trying to get started with Homebox for the first time. Your task is to navigate through the documentation site, follow the getting started guide, and identify any confusing, broken, or unclear steps.

## Context

- Repository: ${{ github.repository }}
- Working directory: ${{ github.workspace }}
- Documentation directory: ${{ github.workspace }}/docs

## Your Mission

Act as a complete beginner who has never used Homebox before. Build and navigate the documentation site, follow tutorials step-by-step, and document any issues you encounter.

## Step 1: Build and Serve Documentation Site

Navigate to the docs folder and build the documentation site using the steps from docs.yml:

```bash
cd ${{ github.workspace }}/docs
npm install
npm run build
```

Follow the shared **Documentation Server Lifecycle Management** instructions:
1. Start the preview server (section "Starting the Documentation Preview Server")
2. Wait for server readiness (section "Waiting for Server Readiness")

## Step 2: Navigate Documentation as a Noob

Using Playwright, navigate through the documentation site as if you're a complete beginner:

1. **Visit the home page** at http://localhost:4321/en/
    - Take a screenshot
    - Note: Is it immediately clear what this tool does?
    - Note: Can you quickly find the "Get Started" or "Quick Start" link?

2. **Follow the Quick Start Guide** at http://localhost:4321/en/quick-start/
    - Take screenshots of each major section
    - Try to understand each step from a beginner's perspective
    - Questions to consider:
        - Are prerequisites clearly listed?
        - Are installation instructions clear and complete?
        - Are there any assumed knowledge gaps?
        - Do code examples work as shown?
        - Are error messages explained?

3. **Explore Install guide** at http://localhost:4321/en/quick-start/install/
    - Take screenshots of confusing sections
    - Note: Is the workflow format explained clearly?
    - Note: Are there enough examples?

4. **Review Contributing Guide** at http://localhost:4321/en/contributing/
    - Take screenshots if explanations are unclear
    - Note: Can you understand how to adapt examples to your own use case?

## Step 3: Identify Pain Points

As you navigate, specifically look for:

### 🔴 Critical Issues (Block getting started)
- Missing prerequisites or dependencies
- Broken links or 404 pages
- Incomplete or incorrect code examples
- Missing critical information
- Confusing navigation structure
- Steps that don't work as described

### 🟡 Confusing Areas (Slow down learning)
- Unclear explanations
- Too much jargon without definitions
- Lack of examples or context
- Inconsistent terminology
- Assumptions about prior knowledge
- Layout or formatting issues that make content hard to read

### 🟢 Good Stuff (What works well)
- Clear, helpful examples
- Good explanations
- Useful screenshots or diagrams
- Logical flow

## Step 4: Take Screenshots

For each confusing or broken area:
- Take a screenshot showing the issue
- Name the screenshot descriptively (e.g., "confusing-quick-start-step-3.png")
- Note the page URL and specific section

## Step 5: Create Discussion Report

Create a GitHub discussion titled "📚 Documentation Noob Test Report - [Date]" with:

### Summary
- Date of test: [Today's date]
- Pages visited: [List URLs]
- Overall impression: [1-2 sentences as a new user]

### Critical Issues Found
[List any blocking issues with screenshots]

### Confusing Areas
[List confusing sections with explanations and screenshots]

### What Worked Well
[Positive feedback on clear sections]

### Recommendations
- Prioritized suggestions for improving the getting started experience
- Quick wins that would help new users immediately
- Longer-term documentation improvements

### Screenshots
[Embed all relevant screenshots showing issues or confusing areas]

Label the discussion with: `documentation`, `user-experience`, `automated-testing`

## Step 6: Cleanup

Follow the shared **Documentation Server Lifecycle Management** instructions for cleanup (section "Stopping the Documentation Server").

## Guidelines

- **Be genuinely naive**: Don't assume knowledge of Git, GitHub Actions, or AI workflows
- **Document everything**: Even minor confusion points matter
- **Be specific**: "This is confusing" is less helpful than "I don't understand what 'frontmatter' means"
- **Be constructive**: Focus on helping improve the docs, not just criticizing
- **Be thorough but efficient**: Cover key getting started paths without testing every single page
- **Take good screenshots**: Make sure they clearly show the issue

## Success Criteria

You've successfully completed this task if you:
- Navigated at least 5 key documentation pages
- Identified specific pain points with examples
- Provided actionable recommendations
- Created a discussion with clear findings and screenshots
