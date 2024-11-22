# Default Channel Plugin for Mattermost

Automatically adding users and guests to predefined channels.

It is an alternative to the inbuilt `ExperimentalDefaultChannels` feature of Mattermost.
This plugin allows you to define a list of workspaces and channels that serve as default channels.
The builtin way of `ExperimentalDefaultChannels` forces to define the same channels for each workspace.


## Configuration

The plugin can be configured via the Mattermost System Console.

The list of defaults channels contain of a list of workspaces and channels.  
They follow the format: `workspace1:channel1,workspace1:channel2,workspace2:channel1,workspace3:channel4`
