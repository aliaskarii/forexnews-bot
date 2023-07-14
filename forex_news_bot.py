import requests
from bs4 import BeautifulSoup
from telegram import Bot
from telegram import ParseMode

def fetch_forex_news():
    url = "https://www.forexfactory.com/calendar"
    response = requests.get(url)
    soup = BeautifulSoup(response.content, "html.parser")
    news_items = soup.find_all("tr", class_="calendar_row")

    news_list = []
    for item in news_items:
        title = item.find("span", class_="calendar__event-title").text.strip()
        time = item.find("td", class_="calendar__time").text.strip()
        news_list.append(f"{time} - {title}")

    return news_list


def publish_news_to_channel(api_token, channel_id, news):
    bot = Bot(api_token)
    message = "\n\n".join(news)
    bot.send_message(channel_id, text=message, parse_mode=ParseMode.HTML)

if name == "main":
    # Set your API token and channel ID
    api_token = "YOUR_API_TOKEN"
    channel_id = "YOUR_CHANNEL_ID"

    # Fetch Forex Factory news
    news = fetch_forex_news()

    # Publish news to Telegram channel
    publish_news_to_channel(api_token, channel_id, news)
