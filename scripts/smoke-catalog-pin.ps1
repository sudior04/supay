param([string]$BaseUrl = 'http://localhost:8080')
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
if (([uri]$BaseUrl).Host -notin @('localhost','127.0.0.1')) { throw 'This fixture smoke check is restricted to the local Supay stack.' }
Set-Location (Split-Path $PSScriptRoot -Parent)
$DockerDir = Join-Path $env:LOCALAPPDATA 'Programs\DockerDesktop\resources\bin'
if (Test-Path $DockerDir) { $env:Path = "$DockerDir;$env:Path" }
$script:Checks = 0
$fixtureIds = [System.Collections.Generic.List[string]]::new()
$categoryIds = [System.Collections.Generic.List[string]]::new()
$suffix = [guid]::NewGuid().ToString('N')
$fixturePassword = 'Fixture.Auth.Password@2026'

function Check([bool]$Condition, [string]$Label) {
  if (!$Condition) { throw "FAILED: $Label" }
  $script:Checks++
  Write-Host "PASS $Label"
}
function Sql([string]$Service, [string]$Database, [string]$Query) {
  $result = & docker compose exec -T $Service psql -U $Database -d $Database -v ON_ERROR_STOP=1 -Atc $Query
  if ($LASTEXITCODE -ne 0) { throw "SQL fixture operation failed in $Database" }
  return $result
}
function Call {
  param([string]$Method,[string]$Path,$Body=$null,[string]$Token='',[int]$Expected=200,$Web=$null,[string]$PIN='AUTO')
  $parameters = @{ Uri="$BaseUrl/api/v1$Path"; Method=$Method; UseBasicParsing=$true; TimeoutSec=15 }
  if ($null -ne $Body) { $parameters.ContentType='application/json'; $parameters.Body=([System.Text.Encoding]::UTF8.GetBytes(($Body | ConvertTo-Json -Depth 15 -Compress))) }
  if ($Token) { $parameters.Headers=@{Authorization="Bearer $Token"} }
  if($PIN -eq 'AUTO'){$PIN=if($Method -in @('POST','PUT','PATCH','DELETE') -and $Path -match '^/(products|orders)(/|$)'){'012345'}else{''}}
  if($PIN){if(!$parameters.Headers){$parameters.Headers=@{}};$parameters.Headers['X-PIN']=$PIN}
  if ($Web) { $parameters.WebSession=$Web }
  $text=''
  try { $response=Invoke-WebRequest @parameters; $status=[int]$response.StatusCode; $text=$response.Content }
  catch { if (!$_.Exception.Response) { throw }; $status=[int]$_.Exception.Response.StatusCode; $text=$_.ErrorDetails.Message }
  Check ($status -eq $Expected) "$Method $Path => $Expected (actual $status)"
  if ($text) { try { return $text | ConvertFrom-Json } catch { if ($Expected -lt 400) { throw } } }
}
function RegisterFixture([string]$Name) {
  $result=Call POST '/auth/register' @{ name="Smoke $Name"; email="$Name-$suffix@example.invalid"; password=$fixturePassword } '' 201
  $fixtureIds.Add($result.user.id)
  Check ([bool]$result.access_token) 'Registration needs no PIN and issues a normal session'
  return $result
}
function LoginFixture($User,$Web=$null) {
  if (!$Web) { $Web=[Microsoft.PowerShell.Commands.WebRequestSession]::new() }
  $result=Call POST '/auth/login' @{email=$User.user.email;password=$fixturePassword} '' 200 $Web
  Check ([bool]$result.access_token -and !$result.challenge_token) 'Password login issues a session without a PIN challenge'
  $rotated=Call POST '/auth/refresh' $null '' 200 $Web
  Check ($rotated.user.id -eq $result.user.id) 'Refresh works without PIN'
  return $rotated
}

try {
  $health=Invoke-RestMethod "$BaseUrl/healthz" -TimeoutSec 10
  Check ($health.service -eq 'gateway') 'NGINX resolves the current Gateway container'
  $stage=Call GET '/products/status'
  Check ($stage.stage -eq 'catalog') 'Gateway resolves the catalog service'
  $null=Call POST '/auth/register' @{name='Invalid';email="invalid-$suffix@example.invalid";password=$fixturePassword;pin='12345'} '' 400
  $admin=RegisterFixture 'admin'; $seller=RegisterFixture 'merchant'; $other=RegisterFixture 'other'; $customer=RegisterFixture 'customer'
  $null=Sql 'user-db' 'user' "UPDATE users SET role='ADMIN' WHERE id='$($admin.user.id)'; UPDATE users SET role='MERCHANT' WHERE id IN ('$($seller.user.id)','$($other.user.id)');"
  $admin=LoginFixture $admin; $seller=LoginFixture $seller; $other=LoginFixture $other
  $null=Call GET '/categories' $null 'not-an-access-token' 401
  $null=Call POST "/users/$($customer.user.id)/pin/reset" $null $customer.access_token 403

  $null=Call GET '/wallets/me' $null $customer.access_token 200
  $null=Call GET '/products' $null $seller.access_token 200
  $null=Call POST '/products' @{} $seller.access_token 403 -PIN ''
  $null=Call POST '/products' @{} $seller.access_token 409
  $status=Call GET '/users/me/pin' $null $seller.access_token 200
  Check (!$status.pin_configured) 'New accounts can sign in before PIN enrollment'
  $null=Call POST '/users/me/pin' @{current_password=$fixturePassword;pin='012345'} $seller.access_token 204
  $null=Call POST '/users/me/pin' @{current_password=$fixturePassword;pin='012345'} $other.access_token 204
  $null=Call POST '/auth/pin/verify' @{pin='012345'} '' 401
  $verified=Call POST '/auth/pin/verify' @{pin='012345'} $seller.access_token 204
  Check ($null -eq $verified) 'PIN verification does not issue a reusable token'
  $null=Call POST '/products' @{} $seller.access_token 403 -PIN ''
  $null=Call POST '/users/me/pin' @{current_password=$fixturePassword;pin='12345'} $customer.access_token 400
  $null=Call POST '/users/me/pin' @{current_password='WrongPassword123';pin='012345'} $customer.access_token 401
  $null=Call POST '/users/me/pin' @{current_password=$fixturePassword;pin='012345'} $customer.access_token 204
  $null=Call POST '/users/me/pin' @{current_password=$fixturePassword;pin='012345'} $customer.access_token 409
  for ($attempt=1;$attempt -le 5;$attempt++) { $expected=if($attempt -eq 5){429}else{401}; $null=Call POST '/auth/pin/verify' @{pin='111111'} $customer.access_token $expected }
  $customer=LoginFixture $customer
  $null=Call GET '/products' $null $customer.access_token 200
  $null=Call POST '/auth/pin/verify' @{pin='012345'} $customer.access_token 429
  $null=Call POST "/users/$($customer.user.id)/pin/reset" $null $admin.access_token 204
  $customer=LoginFixture $customer
  $status=Call GET '/users/me/pin' $null $customer.access_token 200
  Check (!$status.pin_configured) 'PIN reset does not prevent normal login'
  $null=Call POST '/auth/pin/verify' @{pin='012345'} $customer.access_token 409
  $web=[Microsoft.PowerShell.Commands.WebRequestSession]::new()
  $enrolled=LoginFixture $customer $web
  $null=Call POST '/users/me/pin' @{current_password=$fixturePassword;pin='012345'} $enrolled.access_token 204
  $null=Call PUT '/users/me/pin' @{current_password=$fixturePassword;current_pin='111111';new_pin='654321'} $enrolled.access_token 401
  $null=Call PUT '/users/me/pin' @{current_password=$fixturePassword;current_pin='012345';new_pin='654321'} $enrolled.access_token 204 $web
  $null=Call POST '/auth/refresh' $null '' 401 $web
  $customer=LoginFixture $customer
  $null=Call GET '/wallets/me' $null $customer.access_token 200
  $null=Call POST '/auth/pin/verify' @{pin='654321'} $customer.access_token 204

  $categoryBody=@{name="Smoke $suffix";slug="smoke-$suffix";description='Fixture category'}
  $null=Call POST '/categories' $categoryBody $seller.access_token 403
  $category=(Call POST '/categories' $categoryBody $admin.access_token 201).category
  $categoryIds.Add($category.id)
  $OutputEncoding=[System.Text.UTF8Encoding]::new($false)
  $curlCategoryName='Danh m'+[char]0x1ee5+'c ki'+[char]0x1ec3+'m tra'
  $curlCategory=@{name=$curlCategoryName;slug="curl-$suffix";description='UTF-8'} | ConvertTo-Json -Compress |
    curl.exe -sS "$BaseUrl/api/v1/categories" -H "Authorization: Bearer $($admin.access_token)" -H 'Content-Type: application/json' --data-binary '@-' | ConvertFrom-Json
  Check ([bool]$curlCategory.category.id) 'Category cURL pipeline creates a category'
  $categoryIds.Add($curlCategory.category.id)
  Check ($curlCategory.category.name -eq $curlCategoryName) 'Category cURL pipeline preserves Vietnamese UTF-8'
  $null=Call POST '/categories' $categoryBody $admin.access_token 409
  $null=Call GET "/categories/$($category.id)" $null $customer.access_token 200
  $null=Call GET '/categories?limit=1&offset=0' $null $customer.access_token 200
  $categoryBody.name='Smoke updated'
  $null=Call PUT "/categories/$($category.id)" $categoryBody $admin.access_token 200
  $temporary=(Call POST '/categories' @{name='Temporary';slug="temporary-$suffix";description=''} $admin.access_token 201).category
  $categoryIds.Add($temporary.id)
  $null=Call DELETE "/categories/$($temporary.id)" $null $admin.access_token 204

  $body=@{category_id=$category.id;name="Smoke product $suffix";description='Fixture';status='ACTIVE';variants=@(@{sku='RED-M';name='Red M';attributes=@{color='red';size='M'};price_minor=100000;quantity=20},@{sku='BLUE-L';name='Blue L';attributes=@{color='blue';size='L'};price_minor=120000;quantity=10})}
  $null=Call POST '/products' $body $customer.access_token 403 -PIN '654321'
  $invalid=@{category_id=$category.id;name="Rollback-$suffix";description='Fixture';status='ACTIVE';variants=@(@{sku='FIRST';name='First';attributes=@{size='M'};price_minor=1;quantity=1},@{sku='SECOND';name='Second';attributes=@{size='M'};price_minor=1;quantity=1})}
  $null=Call POST '/products' $invalid $seller.access_token 409
  $rollback=Call GET "/products?mine=true&q=Rollback-$suffix" $null $seller.access_token 200
  Check ($rollback.products.Count -eq 0) 'Variant conflict rolls back product and initial inventory'
  $product=(Call POST '/products' $body $seller.access_token 201).product
  Check ($product.variants.Count -eq 2) 'Product + variants + stock created atomically'
  $variant=$product.variants[0]
  $null=Call PUT "/products/$($product.id)" @{category_id=$category.id;name='Blocked';description='';status='ACTIVE'} $seller.access_token 403 -PIN ''
  $null=Call DELETE "/products/$($product.id)" $null $seller.access_token 403 -PIN ''
  $null=Call PATCH "/products/$($product.id)/variants/$($variant.id)/inventory" @{quantity=25;expected_version=0;reason='No PIN'} $seller.access_token 403 -PIN ''
  $null=Call PUT "/products/$($product.id)" @{category_id=$category.id;name='Blocked';description='';status='ACTIVE'} $seller.access_token 401 -PIN '111111'
  $unchanged=(Call GET "/products/$($product.id)" $null $seller.access_token 200).product
  Check ($unchanged.name -eq $product.name) 'Missing/incorrect action PIN does not mutate a product'
  $null=Call GET "/products/$($product.id)" $null $customer.access_token 200
  $page=Call GET "/products?mine=true&q=$suffix&limit=1" $null $seller.access_token 200
  Check ($page.products.Count -eq 1) 'Product list supports merchant/search/pagination'
  $metadata=@{category_id=$category.id;name='Updated product';description='Updated';status='ACTIVE'}
  $null=Call PUT "/products/$($product.id)" $metadata $other.access_token 403
  $null=Call DELETE "/products/$($product.id)" $null $other.access_token 403
  $null=Call PUT "/products/$($product.id)" $metadata $seller.access_token 200
  $null=Call POST "/products/$($product.id)/variants" @{sku='BAD';name='Bad';attributes=@{size='XS'};price_minor=0;quantity=1} $seller.access_token 400
  $null=Call POST "/products/$($product.id)/variants" @{sku='BAD';name='Bad';attributes=@{size='XS'};price_minor=1.5;quantity=1} $seller.access_token 400
  $null=Call POST "/products/$($product.id)/variants" @{sku='BAD';name='Bad';attributes=@{size='XS'};price_minor=1;quantity=-1} $seller.access_token 400
  $null=Call POST "/products/$($product.id)/variants" @{sku=$variant.sku;name='Duplicate';attributes=@{size='XL'};price_minor=1;quantity=1} $seller.access_token 409
  $null=Call POST "/products/$($product.id)/variants" @{sku='DUP-ATTR';name='Duplicate';attributes=$variant.attributes;price_minor=1;quantity=1} $seller.access_token 409
  $null=Call PUT "/products/$($product.id)/variants/$($variant.id)" @{sku=$variant.sku;name='Edited';attributes=$variant.attributes;price_minor=150000} $seller.access_token 200
  $null=Call PATCH "/products/$($product.id)/variants/$($variant.id)/inventory" @{quantity=25;expected_version=0;reason='Restock'} $other.access_token 403
  $stock=Call PATCH "/products/$($product.id)/variants/$($variant.id)/inventory" @{quantity=25;expected_version=0;reason='Restock'} $seller.access_token 200
  Check ($stock.inventory.version -eq 1) 'Stock version advances'
  $null=Call PATCH "/products/$($product.id)/variants/$($variant.id)/inventory" @{quantity=30;expected_version=0;reason='Stale'} $seller.access_token 409

  Add-Type -AssemblyName System.Net.Http
  $client=[System.Net.Http.HttpClient]::new();$client.Timeout=[TimeSpan]::FromSeconds(15)
  try {
    $requests=@();$tasks=@()
    foreach($quantity in @(30,35)) {
      $request=[System.Net.Http.HttpRequestMessage]::new([System.Net.Http.HttpMethod]::new('PATCH'),"$BaseUrl/api/v1/products/$($product.id)/variants/$($variant.id)/inventory")
      $request.Headers.Authorization=[System.Net.Http.Headers.AuthenticationHeaderValue]::new('Bearer',$seller.access_token)
      $request.Headers.Add('X-PIN','012345')
      $request.Content=[System.Net.Http.StringContent]::new((@{quantity=$quantity;expected_version=1;reason='Concurrent fixture'}|ConvertTo-Json -Compress),[System.Text.Encoding]::UTF8,'application/json')
      $requests+=$request;$tasks+=$client.SendAsync($request)
    }
    $statuses=@($tasks | ForEach-Object { [int]$_.GetAwaiter().GetResult().StatusCode } | Sort-Object)
    Check ($statuses.Count -eq 2 -and $statuses[0] -eq 200 -and $statuses[1] -eq 409) 'Concurrent stock writes: exactly one succeeds'
  } finally { foreach($request in $requests){$request.Dispose()};$client.Dispose() }
  $null=Sql 'commerce-db' 'commerce' "UPDATE inventory SET reserved=5 WHERE variant_id='$($variant.id)';"
  $null=Call PATCH "/products/$($product.id)/variants/$($variant.id)/inventory" @{quantity=4;expected_version=2;reason='Below reserved'} $seller.access_token 409
  $null=Call DELETE "/products/$($product.id)" $null $seller.access_token 409
  $null=Sql 'commerce-db' 'commerce' "UPDATE inventory SET reserved=0 WHERE variant_id='$($variant.id)';"

  $draftBody=@{category_id=$category.id;name='Draft fixture';description='';status='DRAFT';variants=@(@{sku='DRAFT';name='Default';attributes=@{};price_minor=1;quantity=0})}
  $draft=(Call POST '/products' $draftBody $seller.access_token 201).product
  $null=Call GET "/products/$($draft.id)" $null $customer.access_token 404
  $null=Call GET "/products/$($draft.id)" $null $seller.access_token 200
  $null=Call DELETE "/products/$($product.id)/variants/$($product.variants[1].id)" $null $seller.access_token 204
  $null=Call DELETE "/products/$($product.id)/variants/$($variant.id)" $null $seller.access_token 409
  $added=Call POST "/products/$($product.id)/variants" @{sku='NEW';name='New';attributes=@{size='XS'};price_minor=1000;quantity=3} $seller.access_token 201
  Check ($added.product.variants.Count -eq 2) 'Adding a variant also provisions inventory'
  $null=Call DELETE "/categories/$($category.id)" $null $admin.access_token 409
  $null=Call DELETE "/products/$($product.id)" $null $seller.access_token 204
  $null=Call GET "/products/$($product.id)" $null $customer.access_token 404
  $null=Call DELETE "/categories/$($category.id)" $null $admin.access_token 409
  $audits=Sql 'commerce-db' 'commerce' "SELECT count(*) FROM inventory_adjustments WHERE variant_id='$($variant.id)';"
  Check ([int]$audits -ge 3) 'Inventory audit includes initial stock, update and concurrent write'
  Write-Host "Passed $script:Checks checks; no PINs, passwords or tokens printed."
}
finally {
  if($fixtureIds.Count -gt 0){
    foreach($id in $fixtureIds){if($id -notmatch '^[0-9a-f-]{36}$'){throw 'Unsafe fixture ID during cleanup'}}
    $ids=($fixtureIds|ForEach-Object{"'$_'"}) -join ','
    $variantQuery="SELECT v.id FROM product_variants v JOIN products p ON p.id=v.product_id WHERE p.merchant_id IN ($ids)"
    $null=Sql 'commerce-db' 'commerce' "BEGIN; DELETE FROM inventory_adjustments WHERE variant_id IN ($variantQuery); DELETE FROM inventory WHERE variant_id IN ($variantQuery); DELETE FROM product_variants WHERE product_id IN (SELECT id FROM products WHERE merchant_id IN ($ids)); DELETE FROM products WHERE merchant_id IN ($ids); COMMIT;"
    $null=Sql 'wallet-db' 'wallet' "DELETE FROM wallet_accounts WHERE owner_id IN ($ids);"
    $null=Sql 'user-db' 'user' "BEGIN; DELETE FROM password_audit_events WHERE actor_user_id IN ($ids) OR target_user_id IN ($ids); DELETE FROM pin_audit_events WHERE actor_user_id IN ($ids) OR target_user_id IN ($ids); DELETE FROM users WHERE id IN ($ids); COMMIT;"
  }
  if($categoryIds.Count -gt 0){foreach($id in $categoryIds){if($id -notmatch '^[0-9a-f-]{36}$'){throw 'Unsafe category fixture ID'}};$ids=($categoryIds|ForEach-Object{"'$_'"}) -join ',';$null=Sql 'commerce-db' 'commerce' "DELETE FROM categories WHERE id IN ($ids);"}
  Write-Host 'Temporary fixtures removed; demo accounts were not enrolled or modified.'
}
