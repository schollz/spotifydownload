const puppeteer = require('puppeteer');

(async () => {
  const browser = await puppeteer.launch({
    headless: true
  });
  const CREDS = require('./creds');

  const page = await browser.newPage();

  await page.goto('https://developer.spotify.com/console/get-playlist-tracks/');
  await page.waitFor("#console-form > div.form-group.header-params > div > span > button");
  await page.evaluate(() => {
    document.querySelector( "#console-form > div.form-group.header-params > div > span > button").scrollIntoView();
  });
  await page.click( "#console-form > div.form-group.header-params > div > span > button");
  await page.waitFor(500);
  await page.click("#scope-playlist-read-private");
  await page.waitFor(50);
  await page.click("#scope-playlist-read-collaborative");
  await page.waitFor(50);
  await page.click("#oauthRequestToken");
  await page.waitFor(1000);
  

  // sign in
  await page.click("#login-username");
  await page.waitFor(50);
  await page.keyboard.type(CREDS.username);
  await page.waitFor(50);
  await page.click("#login-password");
  await page.waitFor(50);
  await page.keyboard.type(CREDS.password);
  await page.waitFor(50);

  await page.click("#login-button");
  await page.waitFor(1000);
  const hrefs = await page.evaluate(() => {
    const anchors = document.querySelectorAll('#oauth-input')[0];
    return anchors.value;
  });
  console.log(hrefs);
  await browser.close();
})();