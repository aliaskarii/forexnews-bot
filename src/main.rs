use minreq::{Response, Error, Request};
use pretty_env_logger::env_logger::from_env;
use serde::Deserialize;
use teloxide::{prelude::*};

#[derive(Debug, Deserialize)]
struct NewsItem {
    title: String,
    #[allow(dead_code)]
    country: String,
    #[allow(dead_code)]
    date: String,
    impact: String,
    #[allow(dead_code)]
    forecast: String,
    #[allow(dead_code)]
    previous: String,
}

const URL: &str = "https://nfs.faireconomy.media/ff_calendar_thisweek.json?version=e7ef4a21d0d488b886475f77d0ca5806";

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


fn main() {
    let json_result:Result<Response, Error> = minreq::get(URL).with_timeout(10).send();
    let response =  match json_result {
        Ok(get) => get , 
        Err(error) => panic!("{}", error),
    };
    let news: Vec<NewsItem> = serde_json::from_str(response.as_str().unwrap()).unwrap();


    news
        .iter()
        .filter(|x| x.impact == "High")
        .for_each(|t| println!("{:#?}",t));

   
}
