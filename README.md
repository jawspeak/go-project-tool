
### Goals

- Manager (or team members) can see:
    - the overall team's code work output
    - adoption / compliance of using Jira. (If someone hasn't touched Jira in a few days, that is a flag)
    - continuous pushing of code changes: prefer small PR's over large.
    - see contentious conversations that some jiras or PR's trigger.
- Sets the expectations of the team culture:
    - leaderboard of Jira activity
    - leaderboard of Stash PR activity (calendar widget view showing trends)
    - shows when people are working trailing 14 days. most active hours in jira / PR's / reviewing, etc.
    - encourages people to work smarter while they work harder, focusing on the high impact work (-:

### Non-Goals

- Don't show people PR's they need to review. Amar's megaman tool does that.


### TODO, next actions

Stash activity is the first net new gain. Get that to be a report I can send out / copy/paste out.

1. Query from Stash to get PR's people are involved in.
    1. iterate over projects in some config that my team cares about: java, web, tarkin, ... (show those stash projects in template w/ template iteration)
    1. lookup PR's for each https://developer.atlassian.com/static/rest/stash/3.11.2/stash-rest.html#idp1808528
    1. and details of those for activity: created by, reviewers https://developer.atlassian.com/static/rest/stash/3.11.2/stash-rest.html#idp1909792 
    1. comments on # comments on each. https://developer.atlassian.com/static/rest/stash/3.11.2/stash-rest.html#idp1613136 
    1. size of code change vs test change
    1. some indication of # of rounds of feedback. e.g. # commits, days PR open, days active comments, days active commits. # commenters (in and out of the team). 
1. Decide what queries make sense for me to use for Jira queries.
    1. unresolved by person
    1. TODO
    1. "actively working on by person" - timeseries last 2 weeks of: created, commented, updated. indicate meta activity of last 1, 2, 3, 7, 30 days comment activity. maybe sparkline of comment velocity? scaled in absolute sense for all the active jiras.
1. Show team view of projects for dashboard around Active Work and upcoming milestones
    1. maybe useful in IP meetings, to highlight if a person has goals for the week, but no actual jiras attached to it.
    1. maybe use a db to create each IP meeting, and assign work to themes of work.
1. Maintain state around the data, perhaps all in a heap that i build myself. sorted by final merge / completion date.
    - this would be used to show a historical trends section, beyond current queries.
    - 