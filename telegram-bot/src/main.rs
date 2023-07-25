use teloxide::{prelude::*};
use pretty_env_logger::env_logger::from_env;


#[tokio::main]
async fn telegram(txt_msg:String) {

        match dotenvy::dotenv() {
            Ok(path) => println!(".env read successfully from {}", path.display()),
            Err(e) => println!("Could not load .env file: {e}"),
        };
   

    pretty_env_logger::init();
    log::info!("Starting News bot...");
    let bot = Bot::from_env();
    teloxide::repl(bot, |bot: Bot, msg: Message| async move {
        bot.send_message(msg.chat.id,txt_msg).await?;
        Ok(())
    })
    .await;
}