use teloxide::{prelude::*};
use pretty_env_logger::env_logger::from_env;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    match dotenvy::dotenv() {
        Ok(path) => println!(".env read successfully from {}", path.display()),
        Err(e) => println!("Could not load .env file: {e}"),
    };
    pretty_env_logger::init();
    log::info!("Starting News bot...");
    let your_id = 196176954;
    let bot = Bot::from_env().auto_send();
    teloxide::repl(bot, |message: Message, bot: AutoSend<Bot>| async move {
        if let Some(text) = message.text() {
            bot.send_message(message.chat.id, text).await?;
        }
        respond(())
    }).await;
    Ok(())
}



   
