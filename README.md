Website scraper 3000


Usage:

```
goscrape [flags] urls...

  -c int
        Concurrency (default 8)
  -f string
        Output format, one of [csv, tab, json] (default "tab")
  -n int
        Maximum amount of urls to scrape
  -o string
        Comma separated list of fields.
                Available fields:
                        url:       the request url
                        path:      the request path
                        query:     the request query params
                        nurls:     amount of scrapable urls on the page
                        status:    the http status code
                        head:      the amount of time it took until headers were received
                        duration:  the total amount of time it took until we received the entire response
                        header.*:  replace * with the header to include in the output
                        meta.*:    replace * with the meta property to include in the output
                        query.*:   replace * with the query param to include in the output
                         (default "status,duration,path,query")
  -t int
        Http timeout in seconds (default 5)
```

Output [tab]

```
200  666.108067ms                                        -
200  152.234315ms  /login                                -
200  516.856092ms  /pricing                              -
200  381.303927ms  /join                                 source=header-home
200  603.319837ms  /marketplace                          -
200  606.098226ms  /business                             -
200  612.261812ms  /features                             -
200  613.576805ms  /                                     -
200  152.719301ms  /open-source                          -
200  678.330653ms  /dashboard                            -
200  171.470853ms  /join                                 source=button-home
200  714.501343ms  /explore                              -
200  154.271834ms  /join                                 plan=business&setup_organization=true&source=business-page
200  153.306646ms  /features/code-review                 -
200  155.688932ms  /features/integrations                -
200  187.044233ms  /features/project-management          -
200  162.383126ms  /personal                             -
404  334.440701ms  / /open-source/stories/freakboy3742   -
404  362.542966ms  / /open-source/stories/ariya          -
404  334.739098ms  / /open-source/stories/kris-nova      -
404  369.249147ms  / /business/customers/mailchimp       -
404  335.887134ms  / /open-source/stories/jessfraz       -
404  348.000441ms  / /business/customers/mapbox          -
404  356.276516ms  / /open-source/stories/yyx990803      -
200  243.531287ms  /about                                -
200  179.194758ms  /business/customers                   -
301  109.882188ms  /site/terms                           -
200  172.297884ms  /about/careers                        -
302  112.644837ms  /site/privacy                         -
200  223.472358ms  /blog                                 -
200  156.380866ms  /contact                              -
200  193.694814ms  /about/press                          -
301  110.846172ms  /security                             -
200  148.04303ms   /login                                return_to=%2Fpricing
200  161.270056ms  /password_reset                       -
200  161.486218ms  /join                                 source=login
200  168.473588ms  /join                                 source=header
200  176.732801ms  /pricing/developer                    -
200  160.051263ms  /join                                 plan=pro&source=button-pricing
200  166.773091ms  /pricing/team                         -
200  145.631146ms  /join                                 plan=business&setup_organization=true&source=button-pricing
200  176.610592ms  /pricing/business-hosted              -
200  173.059243ms  /pricing/business-enterprise          -
200  163.617266ms  /join                                 source=pricing-page-new
200  190.896201ms  /join                                 plan=business_plus&setup_organization=true&source=button-pricing
200  147.093362ms  /login                                return_to=%2Fjoin%3Fsource%3Dheader-home
200  168.431153ms  /join                                 -
200  172.378667ms  /login                                return_to=%2Fmarketplace
200  321.533924ms  /marketplace/blackfire-io             -
200  385.609184ms  /marketplace/codetree                 -

Success:  50
Errors:   0,

Fastest:  109.882188ms
Slowest:  714.501343ms

Mean:     179.194758ms
Average:  276.244959ms

StatusCodes:
        200: 40
        404: 7
        301: 2
        302: 1

```
