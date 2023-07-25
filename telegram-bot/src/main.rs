use teloxide::prelude::{*, ChatId};
#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    match dotenvy::dotenv() {
        Ok(path) => println!(".env read successfully from {}", path.display()),
        Err(e) => println!("Could not load .env file: {e}"),
    };
    pretty_env_logger::init();
    log::info!("Starting News bot...");
    let your_id = ChatId(196176954);
    let text_msg = "test";
    let bot = Bot::from_env();
    bot
    .send_message(your_id, text_msg)
    .protect_content(true) // <-- optional parameter!
    .await?;
    Ok(())
}



   
