# mattermost-twilio-plugin
Mattermost Server Plugin for integrating Twilio

This plugin allows you to receive and reply to text messages sent to your Twilio phone numbers in Mattermost. 

## Installation

1. Build the plugin:

    ```bash
    make dist
    ```

2. Upload the generated `.tar.gz` file to your Mattermost server via **System Console > Plugins > Plugin Management**.

3. Enable the plugin in the Mattermost System Console.

## Configuration

1. Go to **System Console > Plugins > Mattermost Twilio Plugin.
2. Enter your Twilio Account SID, Auth Token, and the team and users you want to use.
3. Save the settings.

## Usage

- You must setup a phone number in Twilio that can use conversations.  
- Use `/twilio number list` to get a list of phone numbers you have setup
- To get twilio to send conversations to mattermost use `/twilio number webhooks setup +1XXXXXXXXXX`.  This can bog things down for a bit if you already have a large number of conversations on that chat service as it sets up a webhook for each conversation.
- Incoming SMS messages to your Twilio number will appear in a designated Mattermost channel. You can rename the channels however you like.
- Reply to messages directly in the channel to send SMS responses via Twilio.

## Requirements

- Mattermost server v10.0 or later (probably works on earlier versions, but no guarantees)
- Twilio account with an active phone number

## License

See [LICENSE](LICENSE) for details.
