use minreq::{Response, Error, Request};
use serde::Deserialize;

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
