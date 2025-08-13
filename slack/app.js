const { App, LogLevel, ExpressReceiver } = require("@slack/bolt");
const bodyParser = require("body-parser");

const port = process.env.PORT || 6000;

const receiver = new ExpressReceiver({
    signingSecret: process.env.SLACK_SIGNING_SECRET,
});
receiver.router.use(bodyParser.json());

const app = new App({
    token: process.env.SLACK_BOT_TOKEN,
    logLevel: LogLevel.DEBUG,
    receiver,
});

receiver.router.post("/v1/webhook", async (req, res) => {
    try {
        if (!req.body) {
            return res.status(400).send("Error: request body is missing");
        }

        const { title_markdown, body_markdown } = req.body;
        if (!title_markdown || !body_markdown) {
            return res
                .status(400)
                .send('Error: missing fields: "title_markdown", or "body_markdown"');
        }

        const payload = req.body.payload;
        if (!payload) {
            return res.status(400).send('Error: missing "payload" field');
        }

        const { user_email, actions } = payload;
        if (!user_email || !actions) {
            return res
                .status(400)
                .send('Error: missing fields: "user_email", "actions"');
        }

        // Get the user ID using Slack API
        const userByEmail = await app.client.users.lookupByEmail({
            email: user_email,
        });

        const slackMessage = {
            channel: userByEmail.user.id,
            text: body_markdown,
            blocks: [
                {
                    type: "header",
                    text: { type: "mrkdwn", text: title_markdown },
                },
                {
                    type: "section",
                    text: { type: "mrkdwn", text: body_markdown },
                },
            ],
        };

        // Add action buttons if they exist
        if (actions && actions.length > 0) {
            slackMessage.blocks.push({
                type: "actions",
                elements: actions.map((action) => ({
                    type: "button",
                    text: { type: "plain_text", text: action.label },
                    url: action.url,
                })),
            });
        }

        // Post message to the user on Slack
        await app.client.chat.postMessage(slackMessage);

        res.status(204).send();
    } catch (error) {
        console.error("Error sending message:", error);
        res.status(500).send();
    }
});


app.action("button_click", async ({ body, ack, say }) => {
    await ack(); 
});

(async () => {
    await app.start(port);
    console.log("⚡️ Coder Slack bot is running!");
})();
