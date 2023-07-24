use serde::Deserialize;
use reqwest::{self, Response};
const JSON_URL: &str = "https://nfs.faireconomy.media/ff_calendar_thisweek.json?version=e7ef4a21d0d488b886475f77d0ca5806";
use std::io;

#[derive(Debug, Deserialize)]
#[serde(rename_all = "PascalCase")]
struct JsonData {
    title: String,
    country: String,
    date: String,
    impact: String,
    forecast: String,
    previous: String,
}

//Json Example News
/*r#"[{
    "title": "G20 Meetings",
    "country": "ALL",
    "date": "2023-07-16T00:45:00-04:00",
    "impact": "Medium",
    "forecast": "",
    "previous": ""
},
{
    "title": "G7 Meetings",
    "country": "ALL",
    "date": "2023-07-16T00:45:00-04:00",
    "impact": "Low",
    "forecast": "",
    "previous": ""
}]"#*/

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let client = reqwest::Client::builder().build()?;

    let res = client
        .get(JSON_URL)
        .send()
        .await?
        .bytes()
        .await?;

    let mut data = res.as_ref();

    let mut f = File::create("week.json")?;

    io::copy(&mut data, &mut f)?;

    Ok(())
}