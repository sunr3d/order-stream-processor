let orderTemplate = '';
let itemTemplate = '';

Promise.all([
    fetch('/order_template.html').then(response => response.text()),
    fetch('/item_template.html').then(response => response.text())
])
.then(([orderTpl, itemTpl]) => {
    orderTemplate = orderTpl;
    itemTemplate = itemTpl;
})
.catch(error => {
    console.error('Ошибка загрузки шаблонов:', error);
});

document.getElementById('searchForm').addEventListener('submit', function(e) {
    e.preventDefault();
    
    const orderId = document.getElementById('orderId').value;
    const resultDiv = document.getElementById('result');
    
    fetch('/order/' + orderId)
        .then(response => response.json())
        .then(data => {
            if (data.order) {
                const order = data.order;
                
                const orderData = {
                    order_uid: order.order_uid,
                    customer_id: order.customer_id,
                    track_number: order.track_number,
                    delivery_service: order.delivery_service,
                    date_created: order.date_created,
                    
                    delivery_name: order.delivery?.name || '-',
                    delivery_phone: order.delivery?.phone || '-',
                    delivery_email: order.delivery?.email || '-',
                    delivery_city: order.delivery?.city || '-',
                    delivery_address: order.delivery?.address || '-',
                    
                    payment_transaction: order.payment?.transaction || '-',
                    payment_provider: order.payment?.provider || '-',
                    payment_goods_total: order.payment?.goods_total || '0',
                    payment_delivery_cost: order.payment?.delivery_cost || '0',
                    payment_custom_fee: order.payment?.custom_fee || '0',
                    payment_amount: order.payment?.amount || '0',
                    payment_currency: order.payment?.currency || '-',
                    payment_payment_dt: formatDate(order.payment?.payment_dt),
                    
                    items_count: order.items?.length || 0,
                    items_html: generateItemsHTML(order.items || [])
                };
                
                const html = renderTemplate(orderTemplate, orderData);
                resultDiv.innerHTML = html;
            } else {
                resultDiv.innerHTML = '<p>Заказ не найден</p>';
            }
        })
        .catch(error => {
            resultDiv.innerHTML = '<p>Ошибка: ' + error.message + '</p>';
        });
});

function renderTemplate(template, data) {
    let result = template;
    
    for (const [key, value] of Object.entries(data)) {
        const placeholder = `{{${key}}}`;
        result = result.replace(new RegExp(placeholder, 'g'), value);
    }
    
    return result;
}

function generateItemsHTML(items) {
    return items.map(item => {
        const itemData = {
            item_name: item.name || 'Без названия',
            item_chrt_id: item.chrt_id || 'артикул не указан',
            item_brand: item.brand || 'бренд не указан',
            item_size: item.size || 'размер не указан',
            item_price: item.price || '0',
            item_sale: item.sale ? item.sale + '%' : 'без скидки',
            item_total_price: item.total_price || '0'
        };
        
        return renderTemplate(itemTemplate, itemData);
    }).join('');
}

function formatDate(dateValue) {
    if (!dateValue) return '-';
    
    try {
        if (typeof dateValue === 'string') {
            return new Date(dateValue).toLocaleString('ru-RU');
        } else if (typeof dateValue === 'number') {
            return new Date(dateValue * 1000).toLocaleString('ru-RU');
        }
    } catch (e) {
        return dateValue.toString();
    }
    
    return dateValue.toString();
}